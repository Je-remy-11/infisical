package com.example.ordersystem.entity;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.*;

class OrderTest {

    private Order createPendingOrder() {
        return Order.create("ORD-001", "user-001", new BigDecimal("99.99"),
                LocalDateTime.now().plusMinutes(30));
    }

    @Nested
    @DisplayName("Order.create factory method")
    class CreateTests {

        @Test
        @DisplayName("Should create order with PENDING status")
        void shouldCreatePendingOrder() {
            LocalDateTime timeoutAt = LocalDateTime.now().plusMinutes(30);
            Order order = Order.create("ORD-001", "user-001", new BigDecimal("99.99"), timeoutAt);

            assertEquals("ORD-001", order.getOrderNo());
            assertEquals("user-001", order.getUserId());
            assertEquals(new BigDecimal("99.99"), order.getAmount());
            assertEquals(OrderStatus.PENDING, order.getStatus());
            assertEquals(timeoutAt, order.getTimeoutAt());
            assertEquals(0, order.getRetryCount());
            assertNotNull(order.getCreatedAt());
            assertNotNull(order.getUpdatedAt());
        }
    }

    @Nested
    @DisplayName("isTimeout check")
    class IsTimeoutTests {

        @Test
        @DisplayName("Should return true when PENDING and timeout has passed")
        void shouldReturnTrueWhenTimeoutPassed() {
            Order order = Order.create("ORD-001", "user-001", new BigDecimal("99.99"),
                    LocalDateTime.now().minusMinutes(1));

            assertTrue(order.isTimeout());
        }

        @Test
        @DisplayName("Should return false when PENDING but timeout not reached")
        void shouldReturnFalseWhenNotTimedOut() {
            Order order = createPendingOrder();

            assertFalse(order.isTimeout());
        }

        @Test
        @DisplayName("Should return false when not PENDING status")
        void shouldReturnFalseWhenNotPending() {
            Order order = createPendingOrder();
            order.setStatus(OrderStatus.CONFIRMED);

            assertFalse(order.isTimeout());
        }

        @Test
        @DisplayName("Should return false when timeoutAt is null")
        void shouldReturnFalseWhenTimeoutAtNull() {
            Order order = createPendingOrder();
            order.setTimeoutAt(null);

            assertFalse(order.isTimeout());
        }
    }

    @Nested
    @DisplayName("markTimeoutCancelled")
    class MarkTimeoutCancelledTests {

        @Test
        @DisplayName("Should change PENDING to TIMEOUT_CANCELLED")
        void shouldChangeToTimeoutCancelled() {
            Order order = createPendingOrder();
            LocalDateTime beforeMark = order.getUpdatedAt();

            order.markTimeoutCancelled();

            assertEquals(OrderStatus.TIMEOUT_CANCELLED, order.getStatus());
            assertNotNull(order.getCancelReason());
            assertTrue(order.getCancelReason().contains("timed out"));
        }

        @Test
        @DisplayName("Should throw when order is not PENDING")
        void shouldThrowWhenNotPending() {
            Order order = createPendingOrder();
            order.setStatus(OrderStatus.CONFIRMED);

            assertThrows(IllegalStateException.class, order::markTimeoutCancelled);
        }

        @Test
        @DisplayName("Should be idempotent-safe: second call on same status throws")
        void shouldBeIdempotentSafe() {
            Order order = createPendingOrder();
            order.markTimeoutCancelled();

            assertThrows(IllegalStateException.class, order::markTimeoutCancelled);
        }
    }

    @Nested
    @DisplayName("incrementRetryCount")
    class RetryCountTests {

        @Test
        @DisplayName("Should increment retry count")
        void shouldIncrementRetryCount() {
            Order order = createPendingOrder();
            assertEquals(0, order.getRetryCount());

            order.incrementRetryCount();
            assertEquals(1, order.getRetryCount());

            order.incrementRetryCount();
            assertEquals(2, order.getRetryCount());
        }
    }
}
