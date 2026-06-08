package com.infisical.ordersystem.order;

import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.Instant;
import java.util.List;

public interface OrderRepository extends JpaRepository<Order, Long> {

    @Query("""
            select o
            from Order o
            where o.status = :status
              and o.timeoutAt <= :now
            order by o.timeoutAt asc
            """)
    List<Order> findTimedOutPendingOrders(@Param("status") OrderStatus status,
                                          @Param("now") Instant now,
                                          Pageable pageable);

    @Modifying(clearAutomatically = true, flushAutomatically = true)
    @Query("""
            update Order o
               set o.status = :cancelledStatus,
                   o.cancelledAt = :now,
                   o.updatedAt = :now
             where o.id = :orderId
               and o.status = :pendingStatus
               and o.timeoutAt <= :now
            """)
    int cancelTimedOutOrder(@Param("orderId") Long orderId,
                            @Param("pendingStatus") OrderStatus pendingStatus,
                            @Param("cancelledStatus") OrderStatus cancelledStatus,
                            @Param("now") Instant now);
}
