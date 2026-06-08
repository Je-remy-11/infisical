package com.example.ordersystem.repository;

import com.example.ordersystem.entity.Order;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;
import org.springframework.boot.test.autoconfigure.orm.jpa.TestEntityManager;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

@DataJpaTest
class OrderRepositoryTest {

    @Autowired
    private TestEntityManager entityManager;

    @Autowired
    private OrderRepository orderRepository;

    @Test
    @DisplayName("Should find expired pending orders")
    void shouldFindExpiredPendingOrders() {
        LocalDateTime now = LocalDateTime.now();

        Order expiredPending = Order.builder()
                .orderNo("ORD001")
                .userId(1L)
                .amount(BigDecimal.valueOf(100))
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(now.minusMinutes(5))
                .createdAt(now.minusMinutes(35))
                .build();
        entityManager.persist(expiredPending);

        Order notExpiredPending = Order.builder()
                .orderNo("ORD002")
                .userId(2L)
                .amount(BigDecimal.valueOf(200))
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(now.plusMinutes(10))
                .createdAt(now.minusMinutes(20))
                .build();
        entityManager.persist(notExpiredPending);

        Order expiredPaid = Order.builder()
                .orderNo("ORD003")
                .userId(3L)
                .amount(BigDecimal.valueOf(300))
                .status(Order.OrderStatus.PAID)
                .timeoutAt(now.minusMinutes(5))
                .createdAt(now.minusMinutes(35))
                .build();
        entityManager.persist(expiredPaid);

        entityManager.flush();

        List<Order> result = orderRepository.findByStatusAndTimeoutAtBefore(
                Order.OrderStatus.PENDING, now);

        assertEquals(1, result.size());
        assertEquals("ORD001", result.get(0).getOrderNo());
    }

    @Test
    @DisplayName("Should update order status only when current status matches (idempotent)")
    void shouldUpdateOnlyWhenCurrentStatusMatches() {
        LocalDateTime now = LocalDateTime.now();
        Order order = Order.builder()
                .orderNo("ORD001")
                .userId(1L)
                .amount(BigDecimal.valueOf(100))
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(now.minusMinutes(5))
                .createdAt(now.minusMinutes(35))
                .build();
        Order savedOrder = entityManager.persist(order);
        entityManager.flush();

        LocalDateTime updatedNow = LocalDateTime.now();

        int firstUpdate = orderRepository.cancelOrderIfPending(
                savedOrder.getId(),
                Order.OrderStatus.PENDING,
                Order.OrderStatus.CANCELLED,
                updatedNow,
                updatedNow
        );

        assertEquals(1, firstUpdate);

        Order updatedOrder = entityManager.find(Order.class, savedOrder.getId());
        assertEquals(Order.OrderStatus.CANCELLED, updatedOrder.getStatus());

        int secondUpdate = orderRepository.cancelOrderIfPending(
                savedOrder.getId(),
                Order.OrderStatus.PENDING,
                Order.OrderStatus.CANCELLED,
                updatedNow,
                updatedNow
        );

        assertEquals(0, secondUpdate);
    }

    @Test
    @DisplayName("Should find order by order number")
    void shouldFindByOrderNo() {
        Order order = Order.builder()
                .orderNo("ORD12345")
                .userId(1L)
                .amount(BigDecimal.valueOf(100))
                .status(Order.OrderStatus.PENDING)
                .timeoutAt(LocalDateTime.now().plusMinutes(30))
                .createdAt(LocalDateTime.now())
                .build();
        entityManager.persist(order);
        entityManager.flush();

        var found = orderRepository.findByOrderNo("ORD12345");

        assertTrue(found.isPresent());
        assertEquals("ORD12345", found.get().getOrderNo());
    }

    @Test
    @DisplayName("Should return empty when order not found by order number")
    void shouldReturnEmptyWhenOrderNotFound() {
        var found = orderRepository.findByOrderNo("NOT_EXISTS");
        assertTrue(found.isEmpty());
    }
}
