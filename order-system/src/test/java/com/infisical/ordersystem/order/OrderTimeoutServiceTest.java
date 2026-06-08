package com.infisical.ordersystem.order;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mockito;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.dao.CannotAcquireLockException;
import org.springframework.data.domain.Pageable;
import org.springframework.retry.annotation.EnableRetry;
import org.springframework.test.context.ContextConfiguration;
import org.springframework.test.context.junit.jupiter.SpringExtension;

import java.time.Clock;
import java.time.Instant;
import java.time.ZoneOffset;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(SpringExtension.class)
@ContextConfiguration(classes = OrderTimeoutServiceTest.TestConfig.class)
class OrderTimeoutServiceTest {

    private static final Instant FIXED_NOW = Instant.parse("2026-06-08T10:15:30Z");

    @Autowired
    private OrderTimeoutService orderTimeoutService;

    @Autowired
    private OrderRepository orderRepository;

    @Test
    void scanAndCancelTimedOutOrdersCancelsExpiredPendingOrders() {
        Order first = new Order("ORD-001");
        first.setId(1L);
        Order second = new Order("ORD-002");
        second.setId(2L);

        when(orderRepository.findTimedOutPendingOrders(eq(OrderStatus.PENDING), eq(FIXED_NOW), any(Pageable.class)))
                .thenReturn(List.of(first, second));
        when(orderRepository.cancelTimedOutOrder(1L, OrderStatus.PENDING, OrderStatus.CANCELLED, FIXED_NOW)).thenReturn(1);
        when(orderRepository.cancelTimedOutOrder(2L, OrderStatus.PENDING, OrderStatus.CANCELLED, FIXED_NOW)).thenReturn(1);

        orderTimeoutService.scanAndCancelTimedOutOrders();

        ArgumentCaptor<Pageable> pageableCaptor = ArgumentCaptor.forClass(Pageable.class);
        verify(orderRepository).findTimedOutPendingOrders(eq(OrderStatus.PENDING), eq(FIXED_NOW), pageableCaptor.capture());
        verify(orderRepository, times(2)).cancelTimedOutOrder(any(), eq(OrderStatus.PENDING), eq(OrderStatus.CANCELLED), eq(FIXED_NOW));
        assertThat(pageableCaptor.getValue().getPageNumber()).isZero();
        assertThat(pageableCaptor.getValue().getPageSize()).isEqualTo(50);
    }

    @Test
    void cancelTimedOutOrderIsIdempotentWhenOrderAlreadyProcessed() {
        when(orderRepository.cancelTimedOutOrder(9L, OrderStatus.PENDING, OrderStatus.CANCELLED, FIXED_NOW)).thenReturn(0);

        boolean cancelled = orderTimeoutService.cancelTimedOutOrder(9L);

        assertThat(cancelled).isFalse();
        verify(orderRepository).cancelTimedOutOrder(9L, OrderStatus.PENDING, OrderStatus.CANCELLED, FIXED_NOW);
    }

    @Test
    void cancelTimedOutOrderRetriesOnTransientException() {
        when(orderRepository.cancelTimedOutOrder(5L, OrderStatus.PENDING, OrderStatus.CANCELLED, FIXED_NOW))
                .thenThrow(new CannotAcquireLockException("lock"))
                .thenReturn(1);

        boolean cancelled = orderTimeoutService.cancelTimedOutOrder(5L);

        assertThat(cancelled).isTrue();
        verify(orderRepository, times(2)).cancelTimedOutOrder(5L, OrderStatus.PENDING, OrderStatus.CANCELLED, FIXED_NOW);
    }

    @Configuration
    @EnableRetry
    static class TestConfig {

        @Bean
        OrderRepository orderRepository() {
            return Mockito.mock(OrderRepository.class);
        }

        @Bean
        Clock clock() {
            return Clock.fixed(FIXED_NOW, ZoneOffset.UTC);
        }

        @Bean
        OrderTimeoutService orderTimeoutService(OrderRepository orderRepository,
                                                Clock clock,
                                                ObjectProvider<OrderTimeoutService> selfProvider) {
            return new OrderTimeoutService(orderRepository, clock, 50, selfProvider);
        }
    }
}
