package com.ordersystem.service;

import com.ordersystem.entity.Order;
import com.ordersystem.entity.OrderStatus;
import com.ordersystem.repository.OrderRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.retry.annotation.Backoff;
import org.springframework.retry.annotation.Retryable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;

@Service
@RequiredArgsConstructor
@Slf4j
public class OrderTimeoutService {

    private final OrderRepository orderRepository;

    @Value("${order.timeout.minutes:30}")
    private int timeoutMinutes;

    @Scheduled(cron = "${order.timeout.schedule.cron:0 * * * * ?}")
    public void scanAndCancelTimeoutOrders() {
        log.info("开始扫描超时订单...");
        LocalDateTime now = LocalDateTime.now();

        try {
            List<Order> timeoutOrders = orderRepository.findTimeoutOrders(OrderStatus.PENDING, now);
            log.info("找到 {} 个超时订单待处理", timeoutOrders.size());

            for (Order order : timeoutOrders) {
                try {
                    cancelTimeoutOrder(order.getId());
                } catch (Exception e) {
                    log.error("取消订单失败: {}", order.getOrderNo(), e);
                }
            }

            log.info("超时订单处理完成");
        } catch (Exception e) {
            log.error("扫描超时订单异常", e);
        }
    }

    @Retryable(
            retryFor = {Exception.class},
            maxAttempts = 3,
            backoff = @Backoff(delay = 1000, multiplier = 2)
    )
    @Transactional
    public void cancelTimeoutOrder(Long orderId) {
        Order order = orderRepository.findByIdForUpdate(orderId)
                .orElse(null);

        if (order == null) {
            log.warn("订单不存在: {}", orderId);
            return;
        }

        if (order.getStatus() != OrderStatus.PENDING) {
            log.warn("订单状态不是PENDING，跳过取消: {}, status: {}", order.getOrderNo(), order.getStatus());
            return;
        }

        if (order.getTimeoutAt().isAfter(LocalDateTime.now())) {
            log.warn("订单还未超时，跳过取消: {}, timeoutAt: {}", order.getOrderNo(), order.getTimeoutAt());
            return;
        }

        order.setStatus(OrderStatus.CANCELLED);
        order.setCancelledAt(LocalDateTime.now());
        orderRepository.save(order);

        log.info("超时订单已自动取消: {}", order.getOrderNo());
    }
}
