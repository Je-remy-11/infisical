package com.example.ordersystem.service;

import com.example.ordersystem.entity.Order;
import com.example.ordersystem.repository.OrderRepository;
import lombok.extern.slf4j.Slf4j;
import org.redisson.api.RBlockingQueue;
import org.redisson.api.RDelayedQueue;
import org.redisson.api.RedissonClient;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.retry.annotation.Backoff;
import org.springframework.retry.annotation.Retryable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import jakarta.annotation.PostConstruct;
import java.io.Serializable;
import java.time.LocalDateTime;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

@Slf4j
@Service
@ConditionalOnProperty(prefix = "order.timeout", name = "mechanism", havingValue = "redisson",
        matchIfMissing = false)
public class RedissonOrderTimeoutService {

    private final RedissonClient redissonClient;
    private final OrderRepository orderRepository;
    private final ExecutorService executorService;

    private static final String DELAYED_QUEUE_NAME = "order:timeout:delayed:queue";

    public record OrderTimeoutMessage(Long orderId, String orderNo) implements Serializable {}

    public RedissonOrderTimeoutService(RedissonClient redissonClient, OrderRepository orderRepository) {
        this.redissonClient = redissonClient;
        this.orderRepository = orderRepository;
        this.executorService = Executors.newSingleThreadExecutor();
    }

    @PostConstruct
    public void init() {
        startConsumer();
    }

    public void scheduleTimeoutCancel(Order order, long delayMinutes) {
        RBlockingQueue<OrderTimeoutMessage> blockingQueue = redissonClient.getBlockingQueue(DELAYED_QUEUE_NAME);
        RDelayedQueue<OrderTimeoutMessage> delayedQueue = redissonClient.getDelayedQueue(blockingQueue);

        long delayMillis = delayMinutes * 60 * 1000;
        OrderTimeoutMessage message = new OrderTimeoutMessage(order.getId(), order.getOrderNo());
        delayedQueue.offer(message, delayMillis);

        log.info("Scheduled order timeout cancellation: orderId={}, orderNo={}, delay={} minutes",
                order.getId(), order.getOrderNo(), delayMinutes);
    }

    private void startConsumer() {
        executorService.submit(() -> {
            RBlockingQueue<OrderTimeoutMessage> blockingQueue =
                    redissonClient.getBlockingQueue(DELAYED_QUEUE_NAME);

            while (!Thread.currentThread().isInterrupted()) {
                try {
                    OrderTimeoutMessage message = blockingQueue.take();
                    processTimeoutOrder(message);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    log.info("Order timeout consumer interrupted");
                    break;
                } catch (Exception e) {
                    log.error("Error processing order timeout message", e);
                }
            }
        });
    }

    @Transactional
    @Retryable(
            retryFor = {Exception.class},
            maxAttempts = 3,
            backoff = @Backoff(delay = 1000, multiplier = 2)
    )
    protected void processTimeoutOrder(OrderTimeoutMessage message) {
        log.info("Processing timeout order: orderId={}, orderNo={}", message.orderId(), message.orderNo());

        Order order = orderRepository.findById(message.orderId()).orElse(null);
        if (order == null) {
            log.warn("Order not found: orderId={}", message.orderId());
            return;
        }

        if (order.getStatus() != Order.OrderStatus.PENDING) {
            log.info("Order is not in PENDING status, skip cancellation: orderId={}, status={}",
                    message.orderId(), order.getStatus());
            return;
        }

        LocalDateTime now = LocalDateTime.now();
        if (order.getTimeoutAt().isAfter(now)) {
            log.info("Order timeout not reached yet, reschedule: orderId={}, timeoutAt={}",
                    message.orderId(), order.getTimeoutAt());
            return;
        }

        int updated = orderRepository.cancelOrderIfPending(
                message.orderId(),
                Order.OrderStatus.PENDING,
                Order.OrderStatus.CANCELLED,
                now,
                now
        );

        if (updated > 0) {
            log.info("Successfully cancelled timeout order: orderId={}, orderNo={}",
                    message.orderId(), message.orderNo());
        } else {
            log.warn("Failed to cancel timeout order (concurrent modification): orderId={}",
                    message.orderId());
        }
    }

    public void shutdown() {
        executorService.shutdown();
    }
}
