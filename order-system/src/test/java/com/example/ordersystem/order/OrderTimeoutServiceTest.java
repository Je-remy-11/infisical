package com.example.ordersystem.order;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.time.Clock;
import java.time.Instant;
import java.time.ZoneOffset;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.UUID;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.dao.TransientDataAccessResourceException;
import org.springframework.data.domain.Pageable;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.TransactionDefinition;
import org.springframework.transaction.TransactionStatus;
import org.springframework.transaction.support.SimpleTransactionStatus;

@ExtendWith(MockitoExtension.class)
class OrderTimeoutServiceTest {

    @Mock
    private OrderRepository orderRepository;

    private PlatformTransactionManager transactionManager;
    private Clock clock;
    private OrderTimeoutService orderTimeoutService;

    @BeforeEach
    void setUp() {
        transactionManager = new NoOpTransactionManager();
        clock = Clock.fixed(Instant.parse("2026-06-08T10:00:00Z"), ZoneOffset.UTC);
        orderTimeoutService = new OrderTimeoutService(orderRepository, transactionManager, clock, 100, 3);
    }

    @Test
    void cancelTimedOutOrdersRetriesTransientFailureAndSkipsAlreadyProcessedOrders() {
        Instant now = clock.instant();
        UUID retriedOrderId = UUID.randomUUID();
        UUID alreadyCancelledOrderId = UUID.randomUUID();

        when(orderRepository.findByStatusAndTimeoutAtLessThanEqualOrderByTimeoutAtAsc(
            eq(OrderStatus.PENDING),
            eq(now),
            any(Pageable.class)
        )).thenReturn(List.of(order(retriedOrderId), order(alreadyCancelledOrderId)), List.of());

        when(orderRepository.cancelIfPending(
            eq(retriedOrderId),
            eq(OrderStatus.PENDING),
            eq(OrderStatus.CANCELLED),
            eq(now),
            eq(now)
        )).thenThrow(new TransientDataAccessResourceException("lock timeout"))
            .thenReturn(1);

        when(orderRepository.cancelIfPending(
            eq(alreadyCancelledOrderId),
            eq(OrderStatus.PENDING),
            eq(OrderStatus.CANCELLED),
            eq(now),
            eq(now)
        )).thenReturn(0);

        int cancelledCount = orderTimeoutService.cancelTimedOutOrders(now);

        assertThat(cancelledCount).isEqualTo(1);
        verify(orderRepository, times(2)).cancelIfPending(
            eq(retriedOrderId),
            eq(OrderStatus.PENDING),
            eq(OrderStatus.CANCELLED),
            eq(now),
            eq(now)
        );
        verify(orderRepository, times(1)).cancelIfPending(
            eq(alreadyCancelledOrderId),
            eq(OrderStatus.PENDING),
            eq(OrderStatus.CANCELLED),
            eq(now),
            eq(now)
        );
    }

    @Test
    void cancelTimedOutOrdersThrowsWhenRetriesExhausted() {
        Instant now = clock.instant();
        UUID orderId = UUID.randomUUID();

        when(orderRepository.findByStatusAndTimeoutAtLessThanEqualOrderByTimeoutAtAsc(
            eq(OrderStatus.PENDING),
            eq(now),
            any(Pageable.class)
        )).thenReturn(List.of(order(orderId)));

        when(orderRepository.cancelIfPending(
            eq(orderId),
            eq(OrderStatus.PENDING),
            eq(OrderStatus.CANCELLED),
            eq(now),
            eq(now)
        )).thenThrow(new TransientDataAccessResourceException("db unavailable"));

        assertThatThrownBy(() -> orderTimeoutService.cancelTimedOutOrders(now))
            .isInstanceOf(IllegalStateException.class)
            .hasMessageContaining(orderId.toString());

        verify(orderRepository, times(3)).cancelIfPending(
            eq(orderId),
            eq(OrderStatus.PENDING),
            eq(OrderStatus.CANCELLED),
            eq(now),
            eq(now)
        );
    }

    @Test
    void orderPrePersistDefaultsTimeoutAtToThirtyMinutesAfterCreation() {
        Instant createdAt = clock.instant();
        Order order = new Order();
        order.setCreatedAt(createdAt);

        order.prePersist();

        assertThat(order.getStatus()).isEqualTo(OrderStatus.PENDING);
        assertThat(order.getTimeoutAt()).isEqualTo(createdAt.plus(30, ChronoUnit.MINUTES));
        assertThat(order.getUpdatedAt()).isEqualTo(createdAt);
    }

    private Order order(UUID orderId) {
        Order order = new Order();
        order.setId(orderId);
        order.setStatus(OrderStatus.PENDING);
        return order;
    }

    private static final class NoOpTransactionManager implements PlatformTransactionManager {

        @Override
        public TransactionStatus getTransaction(TransactionDefinition definition) {
            return new SimpleTransactionStatus();
        }

        @Override
        public void commit(TransactionStatus status) {
        }

        @Override
        public void rollback(TransactionStatus status) {
        }
    }
}
