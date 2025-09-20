-- Create email_templates table
CREATE TABLE email_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    body_html TEXT,
    type VARCHAR(50) NOT NULL,
    language VARCHAR(10) DEFAULT 'vi',
    is_active BOOLEAN DEFAULT TRUE,
    variables JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_type (type),
    INDEX idx_language (language),
    INDEX idx_is_active (is_active),
    INDEX idx_deleted_at (deleted_at)
);

-- Create email_queue table
CREATE TABLE email_queue (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    to_email VARCHAR(255) NOT NULL,
    cc VARCHAR(500),
    bcc VARCHAR(500),
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    body_html TEXT,
    template_id BIGINT UNSIGNED NULL,
    priority INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    error_message TEXT,
    scheduled_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES email_templates(id) ON DELETE SET NULL,
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_created_at (created_at)
);

-- Create email_logs table
CREATE TABLE email_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email_queue_id BIGINT UNSIGNED NOT NULL,
    to_email VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    provider VARCHAR(50),
    provider_id VARCHAR(100),
    sent_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_queue_id) REFERENCES email_queue(id) ON DELETE CASCADE,
    INDEX idx_email_queue_id (email_queue_id),
    INDEX idx_status (status),
    INDEX idx_provider (provider),
    INDEX idx_sent_at (sent_at)
);

-- Create email_configs table
CREATE TABLE email_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    ssl BOOLEAN DEFAULT TRUE,
    tls BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_provider (provider),
    INDEX idx_is_active (is_active)
);

-- Insert default email templates
INSERT INTO email_templates (name, subject, body, body_html, type, language, is_active, variables) VALUES
('order_confirmation_vi', 'Xác nhận đơn hàng #{order_number}', 
'Xin chào {customer_name},\n\nCảm ơn bạn đã đặt hàng tại cửa hàng của chúng tôi!\n\nThông tin đơn hàng:\n- Mã đơn hàng: {order_number}\n- Tổng tiền: {total_amount} VND\n- Phương thức thanh toán: {payment_method}\n\nChúng tôi sẽ xử lý đơn hàng của bạn trong thời gian sớm nhất.\n\nTrân trọng,\nĐội ngũ hỗ trợ',
'<h2>Xác nhận đơn hàng #{order_number}</h2><p>Xin chào <strong>{customer_name}</strong>,</p><p>Cảm ơn bạn đã đặt hàng tại cửa hàng của chúng tôi!</p><h3>Thông tin đơn hàng:</h3><ul><li>Mã đơn hàng: <strong>{order_number}</strong></li><li>Tổng tiền: <strong>{total_amount} VND</strong></li><li>Phương thức thanh toán: <strong>{payment_method}</strong></li></ul><p>Chúng tôi sẽ xử lý đơn hàng của bạn trong thời gian sớm nhất.</p><p>Trân trọng,<br>Đội ngũ hỗ trợ</p>',
'order_confirmation', 'vi', TRUE, '["customer_name", "order_number", "total_amount", "payment_method"]'),

('order_shipped_vi', 'Đơn hàng #{order_number} đã được giao',
'Xin chào {customer_name},\n\nĐơn hàng #{order_number} của bạn đã được giao!\n\nThông tin vận chuyển:\n- Mã vận đơn: {tracking_number}\n- Đơn vị vận chuyển: {shipping_company}\n- Dự kiến nhận hàng: {estimated_delivery}\n\nBạn có thể theo dõi đơn hàng tại: {tracking_url}\n\nTrân trọng,\nĐội ngũ hỗ trợ',
'<h2>Đơn hàng #{order_number} đã được giao</h2><p>Xin chào <strong>{customer_name}</strong>,</p><p>Đơn hàng <strong>#{order_number}</strong> của bạn đã được giao!</p><h3>Thông tin vận chuyển:</h3><ul><li>Mã vận đơn: <strong>{tracking_number}</strong></li><li>Đơn vị vận chuyển: <strong>{shipping_company}</strong></li><li>Dự kiến nhận hàng: <strong>{estimated_delivery}</strong></li></ul><p>Bạn có thể theo dõi đơn hàng tại: <a href="{tracking_url}">{tracking_url}</a></p><p>Trân trọng,<br>Đội ngũ hỗ trợ</p>',
'order_shipped', 'vi', TRUE, '["customer_name", "order_number", "tracking_number", "shipping_company", "estimated_delivery", "tracking_url"]'),

('password_reset_vi', 'Đặt lại mật khẩu',
'Xin chào {customer_name},\n\nBạn đã yêu cầu đặt lại mật khẩu cho tài khoản của mình.\n\nĐể đặt lại mật khẩu, vui lòng nhấp vào liên kết sau:\n{reset_link}\n\nLiên kết này sẽ hết hạn sau {expiry_hours} giờ.\n\nNếu bạn không yêu cầu đặt lại mật khẩu, vui lòng bỏ qua email này.\n\nTrân trọng,\nĐội ngũ hỗ trợ',
'<h2>Đặt lại mật khẩu</h2><p>Xin chào <strong>{customer_name}</strong>,</p><p>Bạn đã yêu cầu đặt lại mật khẩu cho tài khoản của mình.</p><p>Để đặt lại mật khẩu, vui lòng nhấp vào liên kết sau:</p><p><a href="{reset_link}" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Đặt lại mật khẩu</a></p><p>Liên kết này sẽ hết hạn sau <strong>{expiry_hours}</strong> giờ.</p><p>Nếu bạn không yêu cầu đặt lại mật khẩu, vui lòng bỏ qua email này.</p><p>Trân trọng,<br>Đội ngũ hỗ trợ</p>',
'password_reset', 'vi', TRUE, '["customer_name", "reset_link", "expiry_hours"]'),

('welcome_vi', 'Chào mừng bạn đến với cửa hàng!',
'Xin chào {customer_name},\n\nChào mừng bạn đến với cửa hàng của chúng tôi!\n\nTài khoản của bạn đã được tạo thành công.\n\nThông tin tài khoản:\n- Email: {email}\n- Tên: {customer_name}\n\nBạn có thể đăng nhập tại: {login_url}\n\nCảm ơn bạn đã tham gia cùng chúng tôi!\n\nTrân trọng,\nĐội ngũ hỗ trợ',
'<h2>Chào mừng bạn đến với cửa hàng!</h2><p>Xin chào <strong>{customer_name}</strong>,</p><p>Chào mừng bạn đến với cửa hàng của chúng tôi!</p><p>Tài khoản của bạn đã được tạo thành công.</p><h3>Thông tin tài khoản:</h3><ul><li>Email: <strong>{email}</strong></li><li>Tên: <strong>{customer_name}</strong></li></ul><p>Bạn có thể đăng nhập tại: <a href="{login_url}">{login_url}</a></p><p>Cảm ơn bạn đã tham gia cùng chúng tôi!</p><p>Trân trọng,<br>Đội ngũ hỗ trợ</p>',
'welcome', 'vi', TRUE, '["customer_name", "email", "login_url"]');

-- Insert default email config
INSERT INTO email_configs (provider, host, port, username, password, from_email, from_name, ssl, tls, is_active) VALUES
('smtp', 'smtp.gmail.com', 587, 'your-email@gmail.com', 'your-app-password', 'your-email@gmail.com', 'E-commerce Store', TRUE, TRUE, TRUE);
