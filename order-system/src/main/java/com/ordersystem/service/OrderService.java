package com.ordersystem.service;

import com.ordersystem.entity.Order;
import com.ordersystem.entity.OrderStatus;
import com.ordersystem.entity.dto.CreateOrderRequest;
import com.ordersystem.entity.dto.OrderResponse;
import com.ordersystem.repository.OrderRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class OrderService {

    private final OrderRepository orderRepository;

    @Transactional
    public OrderResponse createOrder(CreateOrderRequest request) {
        String orderNo = generateOrderNo();

        Optional<Order> existingOrder = orderRepository.findByOrderNo(orderNo);
        if (existingOrder.isPresent()) {
            log.info("订单已存在，直接返回: {}", orderNo);
            return toOrderResponse(existingOrder.get());
        }

        Order order = Order.builder()
                .orderNo(orderNo)
                .userId(request.getUserId())
                .amount(request.getAmount())
                .status(OrderStatus.PENDING)
                .timeoutAt(LocalDateTime.now().plusMinutes(30))
                .build();

        order = orderRepository.save(order);
        log.info("订单创建成功: {}", orderNo);

        return toOrderResponse(order);
    }

    public OrderResponse getOrderById(Long id) {
        return orderRepository.findById(id)
                .map(this::toOrderResponse)
                .orElse(null);
    }

    public OrderResponse getOrderByOrderNo(String orderNo) {
        return orderRepository.findByOrderNo(orderNo)
                .map(this::toOrderResponse)
                .orElse(null);
    }

    public List<OrderResponse> getAllOrders() {
        return orderRepository.findAll().stream()
                .map(this::toOrderResponse)
                .collect(Collectors.toList());
    }

    @Transactional
    public boolean cancelOrder(Long orderId) {
        Optional<Order> orderOpt = orderRepository.findByIdForUpdate(orderId);
        if (orderOpt.isEmpty()) {
            log.warn("订单不存在: {}", orderId);
            return false;
        }

        Order order = orderOpt.get();
        if (order.getStatus() != OrderStatus.PENDING) {
            log.warn("订单状态不是PENDING，无法取消: {}, status: {}", orderId, order.getStatus());
            return false;
        }

        order.setStatus(OrderStatus.CANCELLED);
        order.setCancelledAt(LocalDateTime.now());
        orderRepository.save(order);
        log.info("订单取消成功: {}", order.getOrderNo());

        return true;
    }

    private String generateOrderNo() {
        return "ORD" + System.currentTimeMillis() + UUID.randomUUID().toString().substring(0, 8).toUpperCase();
    }

    private OrderResponse toOrderResponse(Order order) {
        return OrderResponse.builder()
                .id(order.getId())
                .orderNo(order.getOrderNo())
                .userId(order.getUserId())
                .amount(order.getAmount())
                .status(order.getStatus())
                .timeoutAt(order.getTimeoutAt())
                .cancelledAt(order.getCancelledAt())
                .createdAt(order.getCreatedAt())
                .updatedAt(order.getUpdatedAt())
                .build();
    }
}
