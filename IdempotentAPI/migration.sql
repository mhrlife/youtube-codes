CREATE TABLE account
(
    id      INT AUTO_INCREMENT PRIMARY KEY,
    balance INT NOT NULL DEFAULT 0,
    CHECK ( balance >= 0 )
);

CREATE TABLE stocks
(
    id    INT AUTO_INCREMENT PRIMARY KEY,
    stock INT NOT NULL DEFAULT 0,
    price INT NOT NULL DEFAULT 0,
    CHECK ( stock >= 0 ),
    CHECK ( price >= 0 )
);

CREATE TABLE purchase
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    account_id INT NOT NULL,
    stock_id   INT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES account (id),
    FOREIGN KEY (stock_id) REFERENCES stocks (id),
    INDEX idx_account_stock (account_id, stock_id),
    INDEX idx_stock (stock_id)
);




CREATE TABLE idempotency
(
    id          BINARY(16) PRIMARY KEY,
    purchase_id INT NOT NULL,
    FOREIGN KEY (purchase_id) REFERENCES purchase (id)
);
