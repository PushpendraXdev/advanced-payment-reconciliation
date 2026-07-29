create table audit_logs (
    Id serial primary key,
    Matched_Id int references matches(id),
    Internal_Transaction_Id int references internal_transaction(id),
    Gateway_Transaction_Id int references gateway_transactions(id),
    Action Varchar(50),
    Status VARCHAR(50),
    Created_At timestamp DEFAULT Now()
);