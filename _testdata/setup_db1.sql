DROP DATABASE IF EXISTS test_data;
CREATE DATABASE test_data;
CREATE USER tester WITH PASSWORD 'password';

\c test_data
GRANT ALL ON SCHEMA public TO tester;
