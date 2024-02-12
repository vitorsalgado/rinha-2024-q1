CREATE TYPE Tipo AS ENUM ('d', 'c');

CREATE TABLE clientes (
    id SERIAL PRIMARY KEY,
    limite INT NOT NULL,
    saldo INT NOT NULL
);

CREATE TABLE transacoes (
    id SERIAL PRIMARY KEY,
    cliente_id INT NOT NULL,
    descricao VARCHAR(10) NOT NULL,
    tipo Tipo NOT NULL,
    valor INT NOT NULL,
    realizado_em timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_transacoes_cliente
        FOREIGN KEY (cliente_id)
        REFERENCES clientes(id)
);


INSERT INTO clientes
VALUES 
(1, 100000, 0),
(2, 80000, 0),
(3, 1000000, 0),
(4, 10000000, 0),
(5, 500000, 0);
