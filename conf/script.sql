-- tables

CREATE UNLOGGED TABLE saldos (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    limite INT NOT NULL,
    saldo INT NOT NULL
);

CREATE UNLOGGED TABLE transacoes (
    id SERIAL PRIMARY KEY,
    cliente_id INT NOT NULL,
    descricao VARCHAR(10) NOT NULL,
    tipo CHAR(1) NOT NULL,
    valor INT NOT NULL,
    realizado_em TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT fk_transacoes_saldos
        FOREIGN KEY (cliente_id)
        REFERENCES saldos(id)
);

-- indexes

CREATE INDEX idx_saldos_cliente_id ON saldos (cliente_id);
CREATE INDEX idx_transacaos_cliente_id ON transacoes (cliente_id);

-- functions

CREATE OR REPLACE FUNCTION creditar(fn_cliente_id INT, fn_descricao VARCHAR(10), fn_valor INT)
RETURNS TABLE (fn_res_limite INT, fn_res_saldo_final INT)
AS $$
BEGIN
	PERFORM pg_advisory_xact_lock(fn_cliente_id);

	INSERT INTO transacoes (cliente_id, descricao, tipo, valor) 
        VALUES(fn_cliente_id, fn_descricao, 'c', fn_valor);

	RETURN QUERY
        UPDATE saldos
        SET saldo = saldo + fn_valor
        WHERE cliente_id = fn_cliente_id
        RETURNING limite, saldo;
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION debitar(fn_cliente_id INT, fn_descricao VARCHAR(10), fn_valor INT)
RETURNS TABLE (fn_res_limite INT, fn_res_saldo_final INT, fn_res_code INT)
AS $$
DECLARE v_saldo INT; v_limite INT; v_res_code INT DEFAULT 1;
BEGIN
	PERFORM pg_advisory_xact_lock(fn_cliente_id);

	SELECT limite, saldo
	INTO v_limite, v_saldo
	FROM saldos
	WHERE cliente_id = fn_cliente_id;

	IF v_saldo - fn_valor >= v_limite * -1 THEN 
        INSERT INTO transacoes (cliente_id, descricao, tipo, valor) 
        VALUES(fn_cliente_id, fn_descricao, 'd', fn_valor);
		
		UPDATE saldos
		SET saldo = saldo - fn_valor
		WHERE cliente_id = fn_cliente_id;

        v_res_code := 0;
	END IF;

    RETURN QUERY
        SELECT limite, saldo, v_res_code
        FROM saldos
        WHERE cliente_id = fn_cliente_id;
END;
$$
LANGUAGE plpgsql;

-- insert init data

INSERT INTO saldos (cliente_id, limite, saldo)
VALUES 
(1, 100000, 0),
(2, 80000, 0),
(3, 1000000, 0),
(4, 10000000, 0),
(5, 500000, 0);
