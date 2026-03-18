create table biz_users(
	user_id SERIAL primary key,
	username VARCHAR(50) unique not null,
	role VARCHAR(50) DEFAULT 'USER',
	create_time timestamp DEFAULT CURRENT_TIMESTAMP
);

create table auth_requests(
	request_id VARCHAR(64) primary key,
	user_id INT references biz_users(user_id),
	target_domain VARCHAR(50) not null,
	status VARCHAR(50) default 'PENDING',
	chain_tx_hash VARCHAR(100),
	create_time timestamp default CURRENT_TIMESTAMP,
	update_time timestamp default CURRENT_TIMESTAMP
);

create table business_logs(
	log_id SERIAL primary key,
	request_id VARCHAR(64) references auth_requests(request_id),
	action_type VARCHAR(50) not null,
	description TEXT,
	create_time timestamp default CURRENT_TIMESTAMP
);
