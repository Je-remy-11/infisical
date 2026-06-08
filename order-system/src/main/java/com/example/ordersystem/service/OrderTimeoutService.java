package com.example.ordersystem.service;

import com.example.ordersystem.entity.Order;
import com.example.ordersystem.repository.OrderRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.retry.annotation.Backoff;
import org.springframework.retry.annotation.Retryable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class OrderTimeoutService {

    private final OrderRepository orderRepository;

    private static final int BATCH_SIZE = 100;

    @Scheduled(fixedDelayString = "${order.timeout.scan-interval:60000}")
    public void scanExpiredOrders() {
        log.debug("Starting scheduled scan for expired pending orders...");
        processExpiredOrders();
    }

    @Transactional
    @Retryable(
            retryFor = {Exception.class},
            maxAttempts = 3,
            backoff = @Backoff(delay = 1000, multiplier = 2)
    )
    public void processExpiredOrders() {
        LocalDateTime now = LocalDateTime.now();
        List<Order> expiredOrders = orderRepository
                .findByStatusAndTimeoutAtBefore(Order.OrderStatus.PENDING, now);

        if (expiredOrders.isEmpty()) {
            log.debug("No expired pending orders found");
            return;
        }

        log.info("Found {} expired pending orders to cancel", expiredOrders.size());

        int successCount = 0;
        int failCount = 0;

        for (Order order : expiredOrders) {
            try {
                boolean cancelled = cancelExpiredOrder(order);
                if (cancelled) {
                    successCount++;
                }
            } catch (Exception e) {
                failCount++;
                log.error("Failed to cancel expired order id={}, orderNo={}",
                        order.getId(), order.getOrderNo(), e);
            }
        }

        log.info("Expired orders processing completed. Success: {}, Failed: {}",
                successCount, failCount);
    }

    @Transactional
    protected boolean cancelExpiredOrder(Order order) {
        LocalDateTime now = LocalDateTime.now();

        int updated = orderRepository.cancelOrderIfPending(
                order.getId(),
                Order.OrderStatus.PENDING,
                Order.OrderStatus.CANCELLED,
                now,
                now
        );

        if (updated > 0) {
            log.info("Successfully cancelled expired order id={}, orderNo={}",
                    order.getId(), order.getOrderNo());
            return true;
        } else {
            log.warn("Failed to cancel expired order id={}, orderNo={}, " +
                    "it may have already been processed or status changed",
                    order.getId(), order.getOrderNo());
            return false;
        }
    }

    @Transactional
    @Retryable(
            retryFor = {Exception.class},
            maxAttempts = 3,
            backoff = @Backoff(delay = 500, multiplier = 2)
    )
    public boolean manualCancelExpiredOrder(Long orderId) {
        return orderRepository.findById(orderId)
                .map(this::cancelExpiredOrder)
                .orElse(false);
    }
}
