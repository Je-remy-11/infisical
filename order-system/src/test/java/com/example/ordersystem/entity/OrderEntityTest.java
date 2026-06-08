package com.example.ordersystem.entity;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.*;

class OrderEntityTest {

    @Test
    @DisplayName("Should create order with timeoutAt correctly")
    void shouldCreateOrderWithTimeoutAt() {
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime expectedTimeout = now.plusMinutes(30);

        Order order = Order.builder()
                .id(1L)
                .orderNo("ORD001")
                .userId(123L)
                .amount(BigDecimal.valueOf(299.99))
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(expectedTimeout)
                .createdAt(now)
                .build();

        assertEquals(1L, order.getId());
        assertEquals("ORD001", order.getOrderNo());
        assertEquals(123L, order.getUserId());
        assertEquals(BigDecimal.valueOf(299.99), order.getAmount());
        assertEquals(Order.OrderStatus.PENDING, order.getStatus());
        assertEquals(expectedTimeout, order.getTimeoutAt());
        assertEquals(now, order.getCreatedAt());
    }

    @Test
    @DisplayName("Should update cancelledAt when order is cancelled")
    void shouldSetCancelledAtWhenCancelled() {
        LocalDateTime now = LocalDateTime.now();
        Order order = Order.builder()
                .id(1L)
                .orderNo("ORD001")
                .userId(123L)
                .amount(BigDecimal.valueOf(100))
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(now.plusMinutes(30))
                .createdAt(now)
                .build();

        LocalDateTime cancelledAt = LocalDateTime.now();
        order.setStatus(Order.OrderStatus.CANCELLED);
        order.setCancelledAt(cancelledAt);
        order.setUpdatedAt(cancelledAt);

        assertEquals(Order.OrderStatus.CANCELLED, order.getStatus());
        assertEquals(cancelledAt, order.getCancelledAt());
        assertEquals(cancelledAt, order.getUpdatedAt());
    }

    @Test
    @DisplayName("All OrderStatus enum values should be accessible")
    void shouldHaveAllOrderStatusValues() {
        assertTrue(Order.OrderStatus.valueOf("PENDING") != null);
        assertTrue(Order.OrderStatus.valueOf("PAID") != null);
        assertTrue(Order.OrderStatus.valueOf("SHIPPED") != null);
        assertTrue(Order.OrderStatus.valueOf("COMPLETED") != null);
        assertTrue(Order.OrderStatus.valueOf("CANCELLED") != null);
    }
}
