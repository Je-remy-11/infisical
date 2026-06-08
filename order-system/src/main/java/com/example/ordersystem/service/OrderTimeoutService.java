package com.example.ordersystem.service;

import com.example.ordersystem.config.OrderTimeoutProperties;
import com.example.ordersystem.entity.Order;
import com.example.ordersystem.entity.OrderStatus;
import com.example.ordersystem.repository.OrderRepository;
import java.time.LocalDateTime;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.PageRequest;
import org.springframework.orm.ObjectOptimisticLockingFailureException;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class OrderTimeoutService {

    private static final Logger log = LoggerFactory.getLogger(OrderTimeoutService.class);

    private final OrderRepository orderRepository;
    private final OrderTimeoutProperties properties;

    public OrderTimeoutService(OrderRepository orderRepository,
                               OrderTimeoutProperties properties) {
        this.orderRepository = orderRepository;
        this.properties = properties;
    }

    @Scheduled(cron = "${order.timeout.scan.cron}")
    public void scanAndCancelTimeoutOrders() {
        log.info("Starting timeout order scan...");
        LocalDateTime now = LocalDateTime.now();
        int totalCancelled = 0;
        int totalFailed = 0;

        List<Order> timeoutOrders = orderRepository.findTimeoutOrdersWithRetryCapacity(
                OrderStatus.PENDING,
                now,
                properties.getMaxRetry(),
                PageRequest.of(0, properties.getBatchSize())
        );

        log.info("Found {} timeout orders to process", timeoutOrders.size());

        for (Order order : timeoutOrders) {
            try {
                boolean cancelled = cancelOrderWithIdempotency(order);
                if (cancelled) {
                    totalCancelled++;
                    log.info("Order [{}] cancelled due to timeout, retryCount={}",
                            order.getOrderNo(), order.getRetryCount());
                } else {
                    totalFailed++;
                    log.warn("Order [{}] cancellation skipped (concurrent modification), retryCount={}",
                            order.getOrderNo(), order.getRetryCount());
                }
            } catch (ObjectOptimisticLockingFailureException e) {
                totalFailed++;
                log.warn("Order [{}] optimistic lock conflict, will retry next scan",
                        order.getOrderNo());
            } catch (Exception e) {
                totalFailed++;
                handleRetry(order, e);
            }
        }

        log.info("Timeout scan completed: cancelled={}, failed={}", totalCancelled, totalFailed);
    }

    @Transactional
    public boolean cancelOrderWithIdempotency(Order order) {
        Order freshOrder = orderRepository.findById(order.getId()).orElse(null);
        if (freshOrder == null) {
            log.warn("Order [{}] not found, skipping", order.getId());
            return false;
        }

        if (freshOrder.getStatus() != OrderStatus.PENDING) {
            log.info("Order [{}] status is {}, not PENDING, skipping (idempotent)",
                    freshOrder.getOrderNo(), freshOrder.getStatus());
            return true;
        }

        int updated = orderRepository.cancelOrderWithOptimisticLock(
                freshOrder.getId(),
                OrderStatus.PENDING,
                OrderStatus.TIMEOUT_CANCELLED,
                "Order timed out at " + freshOrder.getTimeoutAt(),
                freshOrder.getVersion()
        );

        if (updated == 0) {
            freshOrder.incrementRetryCount();
            orderRepository.save(freshOrder);
            return false;
        }

        return true;
    }

    private void handleRetry(Order order, Exception e) {
        if (order.getRetryCount() >= properties.getMaxRetry()) {
            log.error("Order [{}] exceeded max retry count {}, giving up. Error: {}",
                    order.getOrderNo(), properties.getMaxRetry(), e.getMessage());
            return;
        }

        try {
            Order freshOrder = orderRepository.findById(order.getId()).orElse(null);
            if (freshOrder != null && freshOrder.getStatus() == OrderStatus.PENDING) {
                freshOrder.incrementRetryCount();
                orderRepository.save(freshOrder);
                log.warn("Order [{}] retry count incremented to {}, will retry next scan",
                        freshOrder.getOrderNo(), freshOrder.getRetryCount());
            }
        } catch (Exception retryException) {
            log.error("Failed to increment retry count for order [{}]: {}",
                    order.getOrderNo(), retryException.getMessage());
        }
    }

    public LocalDateTime calculateTimeoutAt() {
        return LocalDateTime.now().plusMinutes(properties.getMinutes());
    }
}
