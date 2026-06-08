package com.example.ordersystem.repository;

import com.example.ordersystem.entity.Order;
import com.example.ordersystem.entity.OrderStatus;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

@Repository
public interface OrderRepository extends JpaRepository<Order, Long> {

    Optional<Order> findByOrderNo(String orderNo);

    List<Order> findByStatusAndTimeoutAtBefore(OrderStatus status, LocalDateTime timeoutThreshold, Pageable pageable);

    @Modifying
    @Query("UPDATE Order o SET o.status = :targetStatus, o.cancelReason = :reason, o.updatedAt = CURRENT_TIMESTAMP "
            + "WHERE o.id = :id AND o.status = :expectedStatus AND o.version = :version")
    int cancelOrderWithOptimisticLock(
            @Param("id") Long id,
            @Param("expectedStatus") OrderStatus expectedStatus,
            @Param("targetStatus") OrderStatus targetStatus,
            @Param("reason") String reason,
            @Param("version") Long version);

    @Query("SELECT o FROM Order o WHERE o.status = :status AND o.timeoutAt < :threshold AND o.retryCount < :maxRetry")
    List<Order> findTimeoutOrdersWithRetryCapacity(
            @Param("status") OrderStatus status,
            @Param("threshold") LocalDateTime threshold,
            @Param("maxRetry") int maxRetry,
            Pageable pageable);
}
