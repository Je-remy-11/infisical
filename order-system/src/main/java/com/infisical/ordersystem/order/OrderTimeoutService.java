package com.infisical.ordersystem.order;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.dao.TransientDataAccessException;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.retry.annotation.Backoff;
import org.springframework.retry.annotation.Recover;
import org.springframework.retry.annotation.Retryable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Clock;
import java.time.Instant;
import java.util.List;

@Service
public class OrderTimeoutService {

    private static final Logger log = LoggerFactory.getLogger(OrderTimeoutService.class);

    private final OrderRepository orderRepository;
    private final Clock clock;
    private final int batchSize;
    private final ObjectProvider<OrderTimeoutService> selfProvider;

    public OrderTimeoutService(OrderRepository orderRepository, ObjectProvider<OrderTimeoutService> selfProvider) {
        this(orderRepository, Clock.systemUTC(), 200, selfProvider);
    }

    public OrderTimeoutService(OrderRepository orderRepository,
                               Clock clock,
                               int batchSize,
                               ObjectProvider<OrderTimeoutService> selfProvider) {
        this.orderRepository = orderRepository;
        this.clock = clock;
        this.batchSize = batchSize;
        this.selfProvider = selfProvider;
    }

    @Scheduled(fixedDelayString = "${order.timeout.scan-delay-ms:60000}")
    public void scanAndCancelTimedOutOrders() {
        Instant now = Instant.now(clock);
        Pageable pageable = PageRequest.of(0, batchSize);
        List<Order> timedOutOrders = orderRepository.findTimedOutPendingOrders(OrderStatus.PENDING, now, pageable);
        OrderTimeoutService retryableService = selfProvider.getIfAvailable(() -> this);
        timedOutOrders.forEach(order -> retryableService.cancelTimedOutOrder(order.getId()));
    }

    @Transactional
    @Retryable(
            retryFor = TransientDataAccessException.class,
            maxAttemptsExpression = "${order.timeout.retry.max-attempts:3}",
            backoff = @Backoff(delayExpression = "${order.timeout.retry.delay-ms:200}", multiplier = 2.0)
    )
    public boolean cancelTimedOutOrder(Long orderId) {
        Instant now = Instant.now(clock);
        return orderRepository.cancelTimedOutOrder(orderId, OrderStatus.PENDING, OrderStatus.CANCELLED, now) > 0;
    }

    @Recover
    public boolean recover(TransientDataAccessException exception, Long orderId) {
        log.error("Failed to cancel timed out order after retries. orderId={}", orderId, exception);
        return false;
    }
}
