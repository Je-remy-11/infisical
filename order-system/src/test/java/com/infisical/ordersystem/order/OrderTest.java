package com.infisical.ordersystem.order;

import org.junit.jupiter.api.Test;

import java.time.Instant;

import static org.assertj.core.api.Assertions.assertThat;

class OrderTest {

    @Test
    void prePersistInitializesPendingStatusAndTimeoutAt() {
        Instant createdAt = Instant.parse("2026-06-08T09:00:00Z");
        Order order = new Order("ORD-1001");
        order.setStatus(null);
        order.setCreatedAt(createdAt);

        order.prePersist();

        assertThat(order.getStatus()).isEqualTo(OrderStatus.PENDING);
        assertThat(order.getTimeoutAt()).isEqualTo(createdAt.plusSeconds(30 * 60));
        assertThat(order.getUpdatedAt()).isEqualTo(createdAt);
    }
}
