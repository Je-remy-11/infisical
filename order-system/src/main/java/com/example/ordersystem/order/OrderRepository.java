package com.example.ordersystem.order;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

public interface OrderRepository extends JpaRepository<Order, UUID> {

    List<Order> findByStatusAndTimeoutAtLessThanEqualOrderByTimeoutAtAsc(
        OrderStatus status,
        Instant timeoutAt,
        Pageable pageable
    );

    @Modifying(clearAutomatically = true, flushAutomatically = true)
    @Query("""
        update Order o
           set o.status = :cancelled,
               o.updatedAt = :updatedAt
         where o.id = :orderId
           and o.status = :pending
           and o.timeoutAt <= :timeoutAt
        """)
    int cancelIfPending(
        @Param("orderId") UUID orderId,
        @Param("pending") OrderStatus pending,
        @Param("cancelled") OrderStatus cancelled,
        @Param("timeoutAt") Instant timeoutAt,
        @Param("updatedAt") Instant updatedAt
    );
}
