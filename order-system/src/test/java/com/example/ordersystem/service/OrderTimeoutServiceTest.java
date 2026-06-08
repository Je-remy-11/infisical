package com.example.ordersystem.service;

import com.example.ordersystem.config.OrderTimeoutProperties;
import com.example.ordersystem.entity.Order;
import com.example.ordersystem.entity.OrderStatus;
import com.example.ordersystem.repository.OrderRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.domain.Pageable;
import org.springframework.orm.ObjectOptimisticLockingFailureException;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Collections;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderTimeoutServiceTest {

    @Mock
    private OrderRepository orderRepository;

    @Mock
    private OrderTimeoutProperties properties;

    @InjectMocks
    private OrderTimeoutService orderTimeoutService;

    private Order buildTimeoutOrder(Long id, String orderNo, int retryCount) {
        Order order = Order.create(orderNo, "user-001", new BigDecimal("99.99"),
                LocalDateTime.now().minusMinutes(31));
        order.setId(id);
        order.setRetryCount(retryCount);
        order.setVersion(1L);
        return order;
    }

    private Order buildActiveOrder(Long id, String orderNo) {
        Order order = Order.create(orderNo, "user-001", new BigDecimal("99.99"),
                LocalDateTime.now().plusMinutes(30));
        order.setId(id);
        order.setVersion(1L);
        return order;
    }

    @Nested
    @DisplayName("scanAndCancelTimeoutOrders")
    class ScanAndCancelTests {

        @Test
        @DisplayName("Should cancel timeout PENDING orders")
        void shouldCancelTimeoutOrders() {
            Order timeoutOrder = buildTimeoutOrder(1L, "ORD-001", 0);
            when(properties.getMaxRetry()).thenReturn(3);
            when(properties.getBatchSize()).thenReturn(100);
            when(orderRepository.findTimeoutOrdersWithRetryCapacity(
                    eq(OrderStatus.PENDING), any(LocalDateTime.class), eq(3), any(Pageable.class)))
                    .thenReturn(List.of(timeoutOrder));
            when(orderRepository.findById(1L)).thenReturn(Optional.of(timeoutOrder));
            when(orderRepository.cancelOrderWithOptimisticLock(
                    eq(1L), eq(OrderStatus.PENDING), eq(OrderStatus.TIMEOUT_CANCELLED),
                    anyString(), eq(1L)))
                    .thenReturn(1);

            orderTimeoutService.scanAndCancelTimeoutOrders();

            verify(orderRepository).cancelOrderWithOptimisticLock(
                    eq(1L), eq(OrderStatus.PENDING), eq(OrderStatus.TIMEOUT_CANCELLED),
                    anyString(), eq(1L));
        }

        @Test
        @DisplayName("Should skip when no timeout orders found")
        void shouldSkipWhenNoTimeoutOrders() {
            when(properties.getMaxRetry()).thenReturn(3);
            when(properties.getBatchSize()).thenReturn(100);
            when(orderRepository.findTimeoutOrdersWithRetryCapacity(
                    eq(OrderStatus.PENDING), any(LocalDateTime.class), eq(3), any(Pageable.class)))
                    .thenReturn(Collections.emptyList());

            orderTimeoutService.scanAndCancelTimeoutOrders();

            verify(orderRepository, never()).cancelOrderWithOptimisticLock(
                    anyLong(), any(), any(), anyString(), anyLong());
        }

        @Test
        @DisplayName("Should handle optimistic lock conflict gracefully")
        void shouldHandleOptimisticLockConflict() {
            Order timeoutOrder = buildTimeoutOrder(1L, "ORD-001", 0);
            when(properties.getMaxRetry()).thenReturn(3);
            when(properties.getBatchSize()).thenReturn(100);
            when(orderRepository.findTimeoutOrdersWithRetryCapacity(
                    eq(OrderStatus.PENDING), any(LocalDateTime.class), eq(3), any(Pageable.class)))
                    .thenReturn(List.of(timeoutOrder));
            when(orderRepository.findById(1L)).thenReturn(Optional.of(timeoutOrder));
            when(orderRepository.cancelOrderWithOptimisticLock(
                    anyLong(), any(), any(), anyString(), anyLong()))
                    .thenReturn(0);

            orderTimeoutService.scanAndCancelTimeoutOrders();

            verify(orderRepository).save(any(Order.class));
        }
    }

    @Nested
    @DisplayName("cancelOrderWithIdempotency")
    class IdempotencyTests {

        @Test
        @DisplayName("Should return true when order already cancelled (idempotent)")
        void shouldBeIdempotentWhenAlreadyCancelled() {
            Order cancelledOrder = buildTimeoutOrder(1L, "ORD-001", 0);
            cancelledOrder.setStatus(OrderStatus.TIMEOUT_CANCELLED);

            when(orderRepository.findById(1L)).thenReturn(Optional.of(cancelledOrder));

            boolean result = orderTimeoutService.cancelOrderWithIdempotency(cancelledOrder);

            assertTrue(result);
            verify(orderRepository, never()).cancelOrderWithOptimisticLock(
                    anyLong(), any(), any(), anyString(), anyLong());
        }

        @Test
        @DisplayName("Should return true when order already confirmed (idempotent)")
        void shouldBeIdempotentWhenAlreadyConfirmed() {
            Order confirmedOrder = buildActiveOrder(1L, "ORD-001");
            confirmedOrder.setStatus(OrderStatus.CONFIRMED);

            when(orderRepository.findById(1L)).thenReturn(Optional.of(confirmedOrder));

            boolean result = orderTimeoutService.cancelOrderWithIdempotency(confirmedOrder);

            assertTrue(result);
            verify(orderRepository, never()).cancelOrderWithOptimisticLock(
                    anyLong(), any(), any(), anyString(), anyLong());
        }

        @Test
        @DisplayName("Should return false when order not found")
        void shouldReturnFalseWhenOrderNotFound() {
            Order order = buildTimeoutOrder(999L, "ORD-999", 0);
            when(orderRepository.findById(999L)).thenReturn(Optional.empty());

            boolean result = orderTimeoutService.cancelOrderWithIdempotency(order);

            assertFalse(result);
        }

        @Test
        @DisplayName("Should successfully cancel PENDING order")
        void shouldCancelPendingOrder() {
            Order pendingOrder = buildTimeoutOrder(1L, "ORD-001", 0);

            when(orderRepository.findById(1L)).thenReturn(Optional.of(pendingOrder));
            when(orderRepository.cancelOrderWithOptimisticLock(
                    eq(1L), eq(OrderStatus.PENDING), eq(OrderStatus.TIMEOUT_CANCELLED),
                    anyString(), eq(1L)))
                    .thenReturn(1);

            boolean result = orderTimeoutService.cancelOrderWithIdempotency(pendingOrder);

            assertTrue(result);
        }
    }

    @Nested
    @DisplayName("Retry mechanism")
    class RetryTests {

        @Test
        @DisplayName("Should increment retry count on failure")
        void shouldIncrementRetryCountOnFailure() {
            Order timeoutOrder = buildTimeoutOrder(1L, "ORD-001", 0);
            when(properties.getMaxRetry()).thenReturn(3);
            when(properties.getBatchSize()).thenReturn(100);
            when(orderRepository.findTimeoutOrdersWithRetryCapacity(
                    eq(OrderStatus.PENDING), any(LocalDateTime.class), eq(3), any(Pageable.class)))
                    .thenReturn(List.of(timeoutOrder));
            when(orderRepository.findById(1L))
                    .thenReturn(Optional.of(timeoutOrder))
                    .thenReturn(Optional.of(timeoutOrder));
            when(orderRepository.cancelOrderWithOptimisticLock(
                    anyLong(), any(), any(), anyString(), anyLong()))
                    .thenThrow(new ObjectOptimisticLockingFailureException(Order.class, 1L));

            orderTimeoutService.scanAndCancelTimeoutOrders();

            verify(orderRepository).save(any(Order.class));
        }

        @Test
        @DisplayName("Should skip orders exceeding max retry count")
        void shouldSkipOrdersExceedingMaxRetry() {
            when(properties.getMaxRetry()).thenReturn(3);
            when(properties.getBatchSize()).thenReturn(100);
            Order exceededOrder = buildTimeoutOrder(1L, "ORD-001", 4);

            when(orderRepository.findTimeoutOrdersWithRetryCapacity(
                    eq(OrderStatus.PENDING), any(LocalDateTime.class), eq(3), any(Pageable.class)))
                    .thenReturn(Collections.emptyList());

            orderTimeoutService.scanAndCancelTimeoutOrders();

            verify(orderRepository, never()).cancelOrderWithOptimisticLock(
                    anyLong(), any(), any(), anyString(), anyLong());
        }
    }

    @Nested
    @DisplayName("calculateTimeoutAt")
    class CalculateTimeoutTests {

        @Test
        @DisplayName("Should calculate timeout based on configured minutes")
        void shouldCalculateTimeout() {
            when(properties.getMinutes()).thenReturn(30);

            LocalDateTime timeoutAt = orderTimeoutService.calculateTimeoutAt();

            assertNotNull(timeoutAt);
            assertTrue(timeoutAt.isAfter(LocalDateTime.now().plusMinutes(29)));
            assertTrue(timeoutAt.isBefore(LocalDateTime.now().plusMinutes(31)));
        }
    }
}
