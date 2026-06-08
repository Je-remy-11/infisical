package com.example.ordersystem.repository;

import com.example.ordersystem.entity.Order;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Repository
public interface OrderRepository extends JpaRepository<Order, Long> {

    Optional<Order> findByOrderNo(String orderNo);

    List<Order> findByStatusAndTimeoutAtBefore(Order.OrderStatus status, LocalDateTime timeoutAt);

    @Modifying
    @Query("UPDATE Order o SET o.status = :newStatus, o.updatedAt = :updatedAt, o.cancelledAt = :cancelledAt " +
           "WHERE o.id = :orderId AND o.status = :currentStatus")
    int cancelOrderIfPending(@Param("orderId") Long orderId,
                             @Param("currentStatus") Order.OrderStatus currentStatus,
                             @Param("newStatus") Order.OrderStatus newStatus,
                             @Param("updatedAt") LocalDateTime updatedAt,
                             @Param("cancelledAt") LocalDateTime cancelledAt);
}
