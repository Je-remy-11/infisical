package com.example.ordersystem.order;

import java.time.Clock;
import java.time.Instant;
import java.util.List;
import java.util.UUID;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.dao.TransientDataAccessException;
import org.springframework.data.domain.PageRequest;
import org.springframework.orm.ObjectOptimisticLockingFailureException;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.TransactionDefinition;
import org.springframework.transaction.support.TransactionTemplate;

@Service
public class OrderTimeoutService {

    private final OrderRepository orderRepository;
    private final Clock clock;
    private final int batchSize;
    private final int maxRetryAttempts;
    private final TransactionTemplate transactionTemplate;

    public OrderTimeoutService(
        OrderRepository orderRepository,
        PlatformTransactionManager transactionManager,
        @Value("${orders.timeout.batch-size:100}") int batchSize,
        @Value("${orders.timeout.retry.max-attempts:3}") int maxRetryAttempts
    ) {
        this(orderRepository, transactionManager, Clock.systemUTC(), batchSize, maxRetryAttempts);
    }

    OrderTimeoutService(
        OrderRepository orderRepository,
        PlatformTransactionManager transactionManager,
        Clock clock,
        int batchSize,
        int maxRetryAttempts
    ) {
        this.orderRepository = orderRepository;
        this.clock = clock;
        this.batchSize = batchSize;
        this.maxRetryAttempts = maxRetryAttempts;
        this.transactionTemplate = new TransactionTemplate(transactionManager);
        this.transactionTemplate.setPropagationBehavior(TransactionDefinition.PROPAGATION_REQUIRES_NEW);
    }

    @Scheduled(fixedDelayString = "${orders.timeout.scan-delay-ms:60000}")
    public void scanAndCancelTimedOutOrders() {
        cancelTimedOutOrders(clock.instant());
    }

    public int cancelTimedOutOrders(Instant now) {
        int cancelledCount = 0;

        while (true) {
            List<Order> timedOutOrders = orderRepository.findByStatusAndTimeoutAtLessThanEqualOrderByTimeoutAtAsc(
                OrderStatus.PENDING,
                now,
                PageRequest.of(0, batchSize)
            );

            if (timedOutOrders.isEmpty()) {
                return cancelledCount;
            }

            for (Order timedOutOrder : timedOutOrders) {
                if (cancelWithRetry(timedOutOrder.getId(), now)) {
                    cancelledCount++;
                }
            }

            if (timedOutOrders.size() < batchSize) {
                return cancelledCount;
            }
        }
    }

    private boolean cancelWithRetry(UUID orderId, Instant now) {
        RuntimeException lastFailure = null;

        for (int attempt = 1; attempt <= maxRetryAttempts; attempt++) {
            try {
                return Boolean.TRUE.equals(transactionTemplate.execute(status ->
                    orderRepository.cancelIfPending(orderId, OrderStatus.PENDING, OrderStatus.CANCELLED, now, now) > 0
                ));
            } catch (TransientDataAccessException | ObjectOptimisticLockingFailureException ex) {
                lastFailure = ex;
            }
        }

        throw new IllegalStateException("Failed to cancel timed out order after retries: " + orderId, lastFailure);
    }
}
