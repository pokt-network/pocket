DO $$ DECLARE
    schemarecord RECORD;
BEGIN
    FOR schemarecord IN (SELECT schema_name
                         FROM information_schema.schemata
                         WHERE schema_name LIKE 'node%')
LOOP
    EXECUTE 'DROP SCHEMA IF EXISTS ' || quote_ident(schemarecord.schema_name) || ' CASCADE';
END LOOP;
END $$;