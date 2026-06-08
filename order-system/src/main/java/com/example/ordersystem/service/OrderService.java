package com.example.ordersystem.service;

import com.example.ordersystem.entity.Order;
import com.example.ordersystem.repository.OrderRepository;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.UUID;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class OrderService {

    private final OrderRepository orderRepository;
    private final OrderTimeoutService orderTimeoutService;

    public OrderService(OrderRepository orderRepository,
                        OrderTimeoutService orderTimeoutService) {
        this.orderRepository = orderRepository;
        this.orderTimeoutService = orderTimeoutService;
    }

    @Transactional
    public Order createOrder(String userId, BigDecimal amount) {
        String orderNo = generateOrderNo();
        LocalDateTime timeoutAt = orderTimeoutService.calculateTimeoutAt();
        Order order = Order.create(orderNo, userId, amount, timeoutAt);
        return orderRepository.save(order);
    }

    @Transactional(readOnly = true)
    public Order getOrderByOrderNo(String orderNo) {
        return orderRepository.findByOrderNo(orderNo)
                .orElseThrow(() -> new IllegalArgumentException("Order not found: " + orderNo));
    }

    private String generateOrderNo() {
        return "ORD-" + UUID.randomUUID().toString().replace("-", "").substring(0, 16).toUpperCase();
    }
}
