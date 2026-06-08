package com.example.ordersystem.order;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.UUID;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.Id;
import jakarta.persistence.Index;
import jakarta.persistence.PrePersist;
import jakarta.persistence.PreUpdate;
import jakarta.persistence.Table;

@Entity
@Table(
    name = "orders",
    indexes = {
        @Index(name = "idx_orders_status_timeout_at", columnList = "status, timeout_at")
    }
)
public class Order {

    private static final long DEFAULT_TIMEOUT_MINUTES = 30L;

    @Id
    @GeneratedValue
    private UUID id;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 32)
    private OrderStatus status = OrderStatus.PENDING;

    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    @Column(name = "timeout_at", nullable = false)
    private Instant timeoutAt;

    public UUID getId() {
        return id;
    }

    public void setId(UUID id) {
        this.id = id;
    }

    public OrderStatus getStatus() {
        return status;
    }

    public void setStatus(OrderStatus status) {
        this.status = status;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Instant createdAt) {
        this.createdAt = createdAt;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(Instant updatedAt) {
        this.updatedAt = updatedAt;
    }

    public Instant getTimeoutAt() {
        return timeoutAt;
    }

    public void setTimeoutAt(Instant timeoutAt) {
        this.timeoutAt = timeoutAt;
    }

    @PrePersist
    void prePersist() {
        Instant now = createdAt == null ? Instant.now() : createdAt;
        createdAt = now;
        updatedAt = updatedAt == null ? now : updatedAt;
        status = status == null ? OrderStatus.PENDING : status;
        timeoutAt = timeoutAt == null ? now.plus(DEFAULT_TIMEOUT_MINUTES, ChronoUnit.MINUTES) : timeoutAt;
    }

    @PreUpdate
    void preUpdate() {
        updatedAt = Instant.now();
    }
}
