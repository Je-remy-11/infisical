package com.ordersystem.service;

import com.ordersystem.entity.Order;
import com.ordersystem.entity.OrderStatus;
import com.ordersystem.entity.dto.CreateOrderRequest;
import com.ordersystem.entity.dto.OrderResponse;
import com.ordersystem.repository.OrderRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderServiceTest {

    @Mock
    private OrderRepository orderRepository;

    @InjectMocks
    private OrderService orderService;

    private Order testOrder;

    @BeforeEach
    void setUp() {
        testOrder = Order.builder()
                .id(1L)
                .orderNo("ORD1234567890ABCDEF")
                .userId(100L)
                .amount(new BigDecimal("99.99"))
                .status(OrderStatus.PENDING)
                .timeoutAt(LocalDateTime.now().plusMinutes(30))
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();
    }

    @Test
    void testCreateOrder_Success() {
        CreateOrderRequest request = CreateOrderRequest.builder()
                .userId(100L)
                .amount(new BigDecimal("99.99"))
                .build();

        when(orderRepository.findByOrderNo(anyString())).thenReturn(Optional.empty());
        when(orderRepository.save(any(Order.class))).thenReturn(testOrder);

        OrderResponse response = orderService.createOrder(request);

        assertNotNull(response);
        assertEquals(testOrder.getOrderNo(), response.getOrderNo());
        assertEquals(testOrder.getUserId(), response.getUserId());
        assertEquals(testOrder.getAmount(), response.getAmount());
        assertEquals(OrderStatus.PENDING, response.getStatus());
        verify(orderRepository, times(1)).save(any(Order.class));
    }

    @Test
    void testCreateOrder_Idempotent() {
        CreateOrderRequest request = CreateOrderRequest.builder()
                .userId(100L)
                .amount(new BigDecimal("99.99"))
                .build();

        when(orderRepository.findByOrderNo(anyString())).thenReturn(Optional.of(testOrder));

        OrderResponse response = orderService.createOrder(request);

        assertNotNull(response);
        verify(orderRepository, never()).save(any(Order.class));
    }

    @Test
    void testGetOrderById_Success() {
        when(orderRepository.findById(1L)).thenReturn(Optional.of(testOrder));

        OrderResponse response = orderService.getOrderById(1L);

        assertNotNull(response);
        assertEquals(testOrder.getId(), response.getId());
    }

    @Test
    void testGetOrderById_NotFound() {
        when(orderRepository.findById(1L)).thenReturn(Optional.empty());

        OrderResponse response = orderService.getOrderById(1L);

        assertNull(response);
    }

    @Test
    void testGetAllOrders() {
        Order order2 = Order.builder()
                .id(2L)
                .orderNo("ORD9876543210FEDCBA")
                .userId(200L)
                .amount(new BigDecimal("199.99"))
                .status(OrderStatus.PAID)
                .timeoutAt(LocalDateTime.now().plusMinutes(30))
                .createdAt(LocalDateTime.now())
                .updatedAt(LocalDateTime.now())
                .build();

        when(orderRepository.findAll()).thenReturn(Arrays.asList(testOrder, order2));

        List<OrderResponse> responses = orderService.getAllOrders();

        assertEquals(2, responses.size());
    }

    @Test
    void testCancelOrder_Success() {
        when(orderRepository.findByIdForUpdate(1L)).thenReturn(Optional.of(testOrder));
        when(orderRepository.save(any(Order.class))).thenReturn(testOrder);

        boolean result = orderService.cancelOrder(1L);

        assertTrue(result);
        verify(orderRepository, times(1)).save(any(Order.class));
    }

    @Test
    void testCancelOrder_AlreadyCancelled() {
        testOrder.setStatus(OrderStatus.CANCELLED);
        when(orderRepository.findByIdForUpdate(1L)).thenReturn(Optional.of(testOrder));

        boolean result = orderService.cancelOrder(1L);

        assertFalse(result);
        verify(orderRepository, never()).save(any(Order.class));
    }

    @Test
    void testCancelOrder_NotFound() {
        when(orderRepository.findByIdForUpdate(1L)).thenReturn(Optional.empty());

        boolean result = orderService.cancelOrder(1L);

        assertFalse(result);
    }
}
