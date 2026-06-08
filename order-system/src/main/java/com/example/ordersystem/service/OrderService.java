package com.example.ordersystem.service;

import com.example.ordersystem.entity.Order;
import com.example.ordersystem.repository.OrderRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class OrderService {

    private final OrderRepository orderRepository;

    private static final int ORDER_TIMEOUT_MINUTES = 30;

    @Transactional
    public Order createOrder(Long userId, BigDecimal amount) {
        LocalDateTime now = LocalDateTime.now();
        Order order = Order.builder()
                .orderNo(generateOrderNo())
                .userId(userId)
                .amount(amount)
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(now.plusMinutes(ORDER_TIMEOUT_MINUTES))
                .createdAt(now)
                .updatedAt(now)
                .build();
        return orderRepository.save(order);
    }

    @Transactional(readOnly = true)
    public Optional<Order> getOrderById(Long id) {
        return orderRepository.findById(id);
    }

    @Transactional(readOnly = true)
    public Optional<Order> getOrderByOrderNo(String orderNo) {
        return orderRepository.findByOrderNo(orderNo);
    }

    @Transactional(readOnly = true)
    public List<Order> getAllOrders() {
        return orderRepository.findAll();
    }

    @Transactional
    public Order payOrder(Long orderId) {
        Order order = orderRepository.findById(orderId)
                .orElseThrow(() -> new RuntimeException("Order not found"));
        if (order.getStatus() != Order.OrderStatus.PENDING) {
            throw new RuntimeException("Order status is not PENDING");
        }
        order.setStatus(Order.OrderStatus.PAID);
        order.setUpdatedAt(LocalDateTime.now());
        return orderRepository.save(order);
    }

    private String generateOrderNo() {
        return "ORD" + System.currentTimeMillis() + UUID.randomUUID().toString().substring(0, 8).toUpperCase();
    }
}
