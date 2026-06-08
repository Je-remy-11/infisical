package com.example.ordersystem.service;

import com.example.ordersystem.entity.Order;
import com.example.ordersystem.entity.OrderStatus;
import com.example.ordersystem.repository.OrderRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderServiceTest {

    @Mock
    private OrderRepository orderRepository;

    @Mock
    private OrderTimeoutService orderTimeoutService;

    @InjectMocks
    private OrderService orderService;

    @Nested
    @DisplayName("createOrder")
    class CreateOrderTests {

        @Test
        @DisplayName("Should create order with timeout calculated from OrderTimeoutService")
        void shouldCreateOrderWithTimeout() {
            LocalDateTime timeoutAt = LocalDateTime.now().plusMinutes(30);
            when(orderTimeoutService.calculateTimeoutAt()).thenReturn(timeoutAt);
            when(orderRepository.save(any(Order.class))).thenAnswer(invocation -> {
                Order order = invocation.getArgument(0);
                order.setId(1L);
                return order;
            });

            Order result = orderService.createOrder("user-001", new BigDecimal("199.99"));

            assertNotNull(result);
            assertEquals("user-001", result.getUserId());
            assertEquals(new BigDecimal("199.99"), result.getAmount());
            assertEquals(OrderStatus.PENDING, result.getStatus());
            assertEquals(timeoutAt, result.getTimeoutAt());
            verify(orderRepository).save(any(Order.class));
        }
    }

    @Nested
    @DisplayName("getOrderByOrderNo")
    class GetOrderTests {

        @Test
        @DisplayName("Should return order when found")
        void shouldReturnOrderWhenFound() {
            Order order = Order.create("ORD-001", "user-001", new BigDecimal("99.99"),
                    LocalDateTime.now().plusMinutes(30));
            order.setId(1L);
            when(orderRepository.findByOrderNo("ORD-001")).thenReturn(Optional.of(order));

            Order result = orderService.getOrderByOrderNo("ORD-001");

            assertNotNull(result);
            assertEquals("ORD-001", result.getOrderNo());
        }

        @Test
        @DisplayName("Should throw when order not found")
        void shouldThrowWhenNotFound() {
            when(orderRepository.findByOrderNo("NOT-EXIST")).thenReturn(Optional.empty());

            assertThrows(IllegalArgumentException.class,
                    () -> orderService.getOrderByOrderNo("NOT-EXIST"));
        }
    }
}
