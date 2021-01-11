CREATE DATABASE MASK;

CREATE TABLE MASK.USER(
    id int(11) NOT NULL AUTO_INCREMENT,
    pid varchar(10) NOT NULL,
    name varchar(100) NOT NULL,
    passwd varchar(256) NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE MASK.STORE(
    id int(11) NOT NULL AUTO_INCREMENT,
    name varchar(100) NOT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE MASK.INVENTORY(
    id int(11) NOT NULL AUTO_INCREMENT,
    store_id int NOT NULL,
    date DATE NOT NULL,
    stock int(11) NOT NULL,
    PRIMARY KEY(id),
    FOREIGN KEY(store_id) REFERENCES STORE(id)
);

CREATE TABLE MASK.ORDER(
    id int(11) NOT NULL AUTO_INCREMENT,
    user_id int(11) NOT NULL,
    inventory_id int(11) NOT NULL,
    pick_up boolean NOT NULL,
    PRIMARY KEY(id),
    FOREIGN KEY(user_id) REFERENCES USER(id),
    FOREIGN KEY(inventory_id) REFERENCES INVENTORY(id)
);

CREATE USER 'test'@'%' IDENTIFIED BY 'test';
GRANT ALL PRIVILEGES ON MASK.* TO 'test'@'%';
FLUSH PRIVILEGES;