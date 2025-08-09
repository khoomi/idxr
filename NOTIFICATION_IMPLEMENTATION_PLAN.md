# Notification System Implementation Plan

## Overview
This document outlines the complete implementation plan for the Khoomi API notification system. The system will handle both user and shop notifications, supporting various notification types, delivery methods, and real-time updates.

## Current State Analysis

### Issues Identified
1. **Missing Core Model**: The `models.Notification` struct referenced in `notificationservice.go` doesn't exist
2. **Incomplete Service**: `NotificationServiceImpl` has undefined `notificationCollection` field  
3. **Partial Implementation**: Async notification methods are just TODO placeholders
4. **Disconnected Flow**: Review/Cart controllers call notification methods but they're not fully implemented

### Existing Components
- Shop notification system (partially implemented)
- Shop notification settings
- Email service with notification templates
- Notification service injected into controllers
- Cache invalidation for notifications

## Implementation Phases

### Phase 1: Create Core Notification Models ✅ (Completed)
1. **Create UserNotification model** in `pkg/models/notification.go`:
   - Basic notification structure for users
   - Notification types enum (order, review, cart, system, etc.)
   - Priority levels (low, medium, high, urgent)
   - Read/unread status tracking
   - Expiry mechanism for auto-cleanup

2. **Fix NotificationService**:
   - Add missing `notificationCollection` field
   - Update constructor to initialize user notification collection
   - Fix CRUD operations to use proper UserNotification model

### Phase 2: Fix NotificationService Implementation ✅ (Completed)
1. **Update `notificationservice.go`**:
   - Implement proper CRUD operations for notifications
   - Add batch operations for marking notifications as read
   - Add pagination support for notification queries
   - Implement notification expiry cleanup

### Phase 3: Implement Async Notification Handlers ✅ (Completed)
1. **Review Notifications**:
   - Fetch review details with listing and shop info
   - Create notification for shop owner
   - Send email if enabled in shop settings
   - Track notification delivery status

2. **Cart Abandonment Notifications**:
   - Check cart items age (older than 24-48 hours)
   - Create reminder notification for user
   - Send email reminder if enabled in user settings
   - Track abandonment metrics

3. **Order Notifications**:
   - New order notifications for shops
   - Order status updates for users (pending, processing, shipped, delivered)
   - Payment confirmations (success/failure)
   - Refund notifications

4. **Inventory Notifications**:
   - Low stock alerts for shop owners
   - Out of stock notifications
   - Restock reminders

### Phase 4: Add User Notification Endpoints ✅ (Completed)
1. **Create user notification routes**:
   ```
   GET    /api/users/:userid/notifications          - Get user notifications (paginated)
   GET    /api/users/:userid/notifications/unread   - Get unread notifications
   PUT    /api/users/:userid/notifications/:id/read - Mark single as read
   PUT    /api/users/:userid/notifications/read-all - Mark all as read
   DELETE /api/users/:userid/notifications/:id      - Delete notification
   DELETE /api/users/:userid/notifications/expired  - Clean expired notifications
   ```

2. **Add UserController notification methods**:
   - `GetUserNotifications()`
   - `GetUnreadUserNotifications()`
   - `MarkUserNotificationAsRead()`
   - `MarkAllUserNotificationsAsRead()`
   - `DeleteUserNotification()`
   - `DeleteExpiredUserNotifications()`

### Phase 5: Integrate with Existing Flows
1. **Hook into existing operations**:
   
   **Order Flow**:
   - Order creation → Shop notification (new order)
   - Payment success → User & Shop notifications
   - Order shipped → User notification
   - Order delivered → User notification
   
   **Review Flow**:
   - Review submission → Shop notification
   - Review approval → User notification
   - Review flagged → Shop notification
   
   **Cart Flow**:
   - Cart abandonment (24h) → User reminder
   - Item back in stock → User notification
   
   **Shop Flow**:
   - New follower → Shop notification
   - Shop announcement → Follower notifications
   - Shop vacation mode → Follower notifications
   
   **Inventory Flow**:
   - Low stock (< threshold) → Shop notification
   - Out of stock → Shop notification
   - Popular item alert → Shop notification

### Phase 6: Add Real-time Support (Optional)
1. **WebSocket/SSE Implementation**:
   - Real-time notification delivery
   - Live notification counter updates
   - Connection management
   - Fallback to polling for older clients

2. **Push Notification Preparation**:
   - Device token management
   - FCM/APNS integration structure
   - Notification preferences per device

## Data Models

### UserNotification
```go
type UserNotification struct {
    ID                 primitive.ObjectID    // Unique identifier
    UserID            primitive.ObjectID    // Recipient user
    Type              NotificationType      // Type of notification
    Title             string               // Notification title
    Message           string               // Notification message
    Priority          NotificationPriority  // Priority level
    IsRead            bool                 // Read status
    ReadAt            *time.Time           // When it was read
    Data              map[string]any       // Additional context data
    RelatedEntityID   *primitive.ObjectID  // Related entity (order, shop, etc.)
    RelatedEntityType string              // Type of related entity
    CreatedAt         time.Time           // Creation timestamp
    ExpiresAt         *time.Time          // Auto-deletion time
}
```

### NotificationType
- OrderCreated
- OrderShipped
- OrderDelivered
- PaymentSuccess
- PaymentFailed
- ReviewReceived
- CartAbandonment
- LowStock
- OutOfStock
- SystemAlert
- SecurityAlert
- PromotionalOffer

### NotificationPriority
- Low
- Medium
- High
- Urgent

## Database Considerations

### Indexes
```javascript
// User notifications collection
db.UserNotification.createIndex({ "user_id": 1, "created_at": -1 })
db.UserNotification.createIndex({ "user_id": 1, "is_read": 1 })
db.UserNotification.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 })

// Shop notifications collection
db.ShopNotification.createIndex({ "shop_id": 1, "created_at": -1 })
db.ShopNotification.createIndex({ "shop_id": 1, "is_read": 1 })
db.ShopNotification.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 })
```

### TTL for Auto-cleanup
- System notifications: 30 days
- Order notifications: 90 days
- Promotional notifications: 7 days
- Security alerts: Never expire

## Implementation Guidelines

### Best Practices
1. **Always check notification preferences** before sending
2. **Batch notifications** when possible to reduce noise
3. **Use appropriate priority levels** to help users filter
4. **Include actionable data** in notification payload
5. **Log all notification events** for debugging
6. **Handle failures gracefully** with retry logic
7. **Implement rate limiting** to prevent notification spam

### Error Handling
- Log failed notification attempts
- Implement exponential backoff for retries
- Provide fallback delivery methods
- Monitor delivery success rates

### Performance Considerations
- Use background workers for async operations
- Implement caching for frequently accessed notifications
- Paginate notification queries
- Archive old notifications instead of deleting

## Testing Strategy

### Unit Tests
- Model validation
- Service method functionality
- Controller endpoint logic

### Integration Tests
- End-to-end notification flow
- Email delivery
- Database operations
- Cache invalidation

### Load Tests
- High volume notification creation
- Concurrent read/write operations
- Notification cleanup performance

## Monitoring & Metrics

### Key Metrics to Track
- Notification delivery rate
- Average delivery time
- Read rate by notification type
- User engagement with notifications
- System performance impact

### Alerting Thresholds
- Delivery failure rate > 5%
- Notification queue size > 1000
- Processing time > 5 seconds
- Database connection errors

## Security Considerations

1. **Access Control**: Users can only access their own notifications
2. **Data Sanitization**: Sanitize all user-generated content in notifications
3. **Rate Limiting**: Prevent notification flooding attacks
4. **Encryption**: Encrypt sensitive data in notification payloads
5. **Audit Logging**: Log all notification-related actions

## Migration Plan

1. **Phase 1**: Deploy core models without breaking changes
2. **Phase 2**: Gradually migrate existing notification code
3. **Phase 3**: Enable new notification types one by one
4. **Phase 4**: Deprecate old notification methods
5. **Phase 5**: Clean up legacy code

## Rollback Strategy

- Feature flags for new notification types
- Database migrations with rollback scripts
- API versioning for backward compatibility
- Monitoring for quick issue detection
- Automated rollback triggers

## Future Enhancements

1. **Machine Learning**: Smart notification timing based on user behavior
2. **Notification Templates**: Customizable templates for different types
3. **Multi-language Support**: Localized notifications
4. **Analytics Dashboard**: Notification performance metrics
5. **A/B Testing**: Test different notification strategies
6. **Smart Bundling**: Intelligently group related notifications
