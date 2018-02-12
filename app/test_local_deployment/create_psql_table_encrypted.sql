CREATE TABLE i2b2demodata.demo_data_encrypted
(
    location_cd character varying(50) COLLATE pg_catalog."default",
    "time" numeric(18,5),
    concept_path character varying(255) COLLATE pg_catalog."default",
    totalnum character varying(88) COLLATE pg_catalog."default"
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE i2b2demodata.demo_data_encrypted
    OWNER to i2b2demodata;
