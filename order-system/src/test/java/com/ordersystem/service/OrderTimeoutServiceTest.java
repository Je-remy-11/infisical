package com.ordersystem.service;

import com.ordersystem.entity.Order;
import com.ordersystem.entity.OrderStatus;
import com.ordersystem.repository.OrderRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Arrays;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderTimeoutServiceTest {

    @Mock
    private OrderRepository orderRepository;

    @InjectMocks
    private OrderTimeoutService orderTimeoutService;

    private Order pendingOrder;
    private Order timeoutOrder;

    @BeforeEach
    void setUp() {
        pendingOrder = Order.builder()
                .id(1L)
                .orderNo("ORD1234567890ABCDEF")
                .userId(100L)
                .amount(new BigDecimal("99.99"))
                .status(OrderStatus.PENDING)
                .timeoutAt(LocalDateTime.now().plusMinutes(30))
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();

        timeoutOrder = Order.builder()
                .id(2L)
                .orderNo("ORD9876543210FEDCBA")
                .userId(200L)
                .amount(new BigDecimal("199.99"))
                .status(OrderStatus.PENDING)
                .timeoutAt(LocalDateTime.now().minusMinutes(5))
                .createdAt(LocalDateTime.now().minusMinutes(35))
                .updatedAt(LocalDateTime.now().minusMinutes(35))
                .build();
    }

    @Test
    void testScanAndCancelTimeoutOrders() {
        when(orderRepository.findTimeoutOrders(eq(OrderStatus.PENDING), any(LocalDateTime.class)))
                .thenReturn(Arrays.asList(timeoutOrder));
        when(orderRepository.findByIdForUpdate(2L)).thenReturn(Optional.of(timeoutOrder));
        when(orderRepository.save(any(Order.class))).thenReturn(timeoutOrder);

        orderTimeoutService.scanAndCancelTimeoutOrders();

        verify(orderRepository, times(1)).save(any(Order.class));
    }

    @Test
    void testCancelTimeoutOrder_Success() {
        when(orderRepository.findByIdForUpdate(2L)).thenReturn(Optional.of(timeoutOrder));
        when(orderRepository.save(any(Order.class))).thenReturn(timeoutOrder);

        orderTimeoutService.cancelTimeoutOrder(2L);

        verify(orderRepository, times(1)).save(any(Order.class));
        assertEquals(OrderStatus.CANCELLED, timeoutOrder.getStatus());
        assertNotNull(timeoutOrder.getCancelledAt());
    }

    @Test
    void testCancelTimeoutOrder_NotPending() {
        timeoutOrder.setStatus(OrderStatus.PAID);
        when(orderRepository.findByIdForUpdate(2L)).thenReturn(Optional.of(timeoutOrder));

        orderTimeoutService.cancelTimeoutOrder(2L);

        verify(orderRepository, never()).save(any(Order.class));
    }

    @Test
    void testCancelTimeoutOrder_NotTimeoutYet() {
        timeoutOrder.setTimeoutAt(LocalDateTime.now().plusMinutes(30));
        when(orderRepository.findByIdForUpdate(2L)).thenReturn(Optional.of(timeoutOrder));

        orderTimeoutService.cancelTimeoutOrder(2L);

        verify(orderRepository, never()).save(any(Order.class));
    }

    @Test
    void testCancelTimeoutOrder_NotFound() {
        when(orderRepository.findByIdForUpdate(2L)).thenReturn(Optional.empty());

        orderTimeoutService.cancelTimeoutOrder(2L);

        verify(orderRepository, never()).save(any(Order.class));
    }

    @Test
    void testCancelTimeoutOrder_Idempotent() {
        timeoutOrder.setStatus(OrderStatus.CANCELLED);
        when(orderRepository.findByIdForUpdate(2L)).thenReturn(Optional.of(timeoutOrder));

        orderTimeoutService.cancelTimeoutOrder(2L);

        verify(orderRepository, never()).save(any(Order.class));
    }
}
