package com.example.ordersystem.service;

import com.example.ordersystem.entity.Order;
import com.example.ordersystem.repository.OrderRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderTimeoutServiceTest {

    @Mock
    private OrderRepository orderRepository;

    private OrderTimeoutService orderTimeoutService;

    @BeforeEach
    void setUp() {
        orderTimeoutService = new OrderTimeoutService(orderRepository);
    }

    @Test
    @DisplayName("Should find and process expired pending orders")
    void shouldProcessExpiredOrders() {
        LocalDateTime now = LocalDateTime.now();
        Order expiredOrder1 = createExpiredOrder(1L, "ORD001", now.minusMinutes(5));
        Order expiredOrder2 = createExpiredOrder(2L, "ORD002", now.minusMinutes(10));

        when(orderRepository.findByStatusAndTimeoutAtBefore(Order.OrderStatus.PENDING, now))
                .thenReturn(List.of(expiredOrder1, expiredOrder2));
        when(orderRepository.cancelOrderIfPending(
                eq(1L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                any(LocalDateTime.class), any(LocalDateTime.class))
        ).thenReturn(1);
        when(orderRepository.cancelOrderIfPending(
                eq(2L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                any(LocalDateTime.class), any(LocalDateTime.class))
        ).thenReturn(1);

        orderTimeoutService.processExpiredOrders();

        verify(orderRepository, times(2)).cancelOrderIfPending(
                anyLong(), any(), any(), any(), any()
        );
    }

    @Test
    @DisplayName("Should do nothing when no expired orders")
    void shouldDoNothingWhenNoExpiredOrders() {
        LocalDateTime now = LocalDateTime.now();
        when(orderRepository.findByStatusAndTimeoutAtBefore(Order.OrderStatus.PENDING, now))
                .thenReturn(List.of());

        orderTimeoutService.processExpiredOrders();

        verify(orderRepository, never()).cancelOrderIfPending(
                anyLong(), any(), any(), any(), any()
        );
    }

    @Test
    @DisplayName("Should return true when successfully cancel expired order")
    void shouldReturnTrueWhenSuccessfullyCancelExpiredOrder() {
        Order order = createExpiredOrder(1L, "ORD001", LocalDateTime.now().minusMinutes(5));

        when(orderRepository.cancelOrderIfPending(
                eq(1L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                any(LocalDateTime.class), any(LocalDateTime.class))
        ).thenReturn(1);

        boolean result = orderTimeoutService.cancelExpiredOrder(order);

        assertTrue(result);
    }

    @Test
    @DisplayName("Should return false when order already cancelled by another process (idempotency)")
    void shouldReturnFalseWhenOrderAlreadyCancelledIdempotency() {
        Order order = createExpiredOrder(1L, "ORD001", LocalDateTime.now().minusMinutes(5));

        when(orderRepository.cancelOrderIfPending(
                eq(1L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                any(LocalDateTime.class), any(LocalDateTime.class))
        ).thenReturn(0);

        boolean result = orderTimeoutService.cancelExpiredOrder(order);

        assertFalse(result);
    }

    @Test
    @DisplayName("Should ensure idempotency with conditional update")
    void shouldEnsureIdempotencyWithConditionalUpdate() {
        Order order = createExpiredOrder(1L, "ORD001", LocalDateTime.now().minusMinutes(5));

        when(orderRepository.cancelOrderIfPending(
                eq(1L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                any(LocalDateTime.class), any(LocalDateTime.class))
        ).thenReturn(1);

        boolean firstResult = orderTimeoutService.cancelExpiredOrder(order);
        when(orderRepository.cancelOrderIfPending(
                eq(1L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                any(LocalDateTime.class), any(LocalDateTime.class))
        ).thenReturn(0);
        boolean secondResult = orderTimeoutService.cancelExpiredOrder(order);

        assertTrue(firstResult);
        assertFalse(secondResult);
        verify(orderRepository, times(2)).cancelOrderIfPending(
                eq(1L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                any(LocalDateTime.class), any(LocalDateTime.class)
        );
    }

    @Test
    @DisplayName("Manual cancel should return false for non-existent order")
    void manualCancelShouldReturnFalseForNonExistentOrder() {
        when(orderRepository.findById(999L)).thenReturn(java.util.Optional.empty());

        boolean result = orderTimeoutService.manualCancelExpiredOrder(999L);

        assertFalse(result);
    }

    @Test
    @DisplayName("Should set correct cancelledAt and updatedAt when cancelling")
    void shouldSetCorrectTimestampsWhenCancelling() {
        Order order = createExpiredOrder(1L, "ORD001", LocalDateTime.now().minusMinutes(5));

        ArgumentCaptor<LocalDateTime> updatedAtCaptor = ArgumentCaptor.forClass(LocalDateTime.class);
        ArgumentCaptor<LocalDateTime> cancelledAtCaptor = ArgumentCaptor.forClass(LocalDateTime.class);

        when(orderRepository.cancelOrderIfPending(
                eq(1L), eq(Order.OrderStatus.PENDING), eq(Order.OrderStatus.CANCELLED),
                updatedAtCaptor.capture(), cancelledAtCaptor.capture())
        ).thenReturn(1);

        LocalDateTime beforeCall = LocalDateTime.now();
        orderTimeoutService.cancelExpiredOrder(order);
        LocalDateTime afterCall = LocalDateTime.now();

        LocalDateTime updatedAt = updatedAtCaptor.getValue();
        LocalDateTime cancelledAt = cancelledAtCaptor.getValue();

        assertTrue(updatedAt.isAfter(beforeCall) || updatedAt.isEqual(beforeCall));
        assertTrue(updatedAt.isBefore(afterCall) || updatedAt.isEqual(afterCall));
        assertEquals(updatedAt, cancelledAt);
    }

    private Order createExpiredOrder(Long id, String orderNo, LocalDateTime timeoutAt) {
        return Order.builder()
                .id(id)
                .orderNo(orderNo)
                .userId(123L)
                .amount(BigDecimal.valueOf(100))
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(timeoutAt)
                .createdAt(LocalDateTime.now().minusMinutes(40))
                .build();
    }
}
