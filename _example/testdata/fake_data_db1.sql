INSERT INTO authors (name, bio)
VALUES
    ('Hemingway', 'Lost Generation boxer'),
    ('Coleridge', 'Lover of opium'),
    ('Johnson', 'Elizabethan poet and playwright'),
    ('Chaucer', 'Vagabond poet'),
    ('de Troyes', 'Provencal Troubador who invented Camelot');

INSERT INTO books (authorID, title, edition, volume, year, wordcount)
VALUES
    ((SELECT id FROM authors WHERE name = 'Hemingway'),
    'The Torrents of Spring', 1, 1, '1926-03-13', 24000),
    ((SELECT id FROM authors WHERE name = 'Hemingway'),
    'The Sun Also Rises', 1, 1, '1926-10-22', 67000),
    ((SELECT id FROM authors WHERE name = 'Hemingway'),
    'A Farewell to Arms', 1, 1, '1929-09-01', 76000),
    ((SELECT id FROM authors WHERE name = 'Hemingway'),
    'For Whom the Bell Tolls', 1, 1, '1940-10-21', 174000),
    ((SELECT id FROM authors WHERE name = 'Chaucer'),
    'The Canterbury Tales', 1, 1, '1400-01-01', 332000),
    ((SELECT id FROM authors WHERE name = 'Coleridge'),
    'The Rime of the Ancient Mariner', 1, 1, '1798-10-04', 7000),
    ((SELECT id FROM authors WHERE name = 'de Troyes'),
    'The Knight of the Cart', 1, 1, '1598-01-01', 73000),
    ((SELECT id FROM authors WHERE name = 'Johnson'),
    'Every Man in His Humour', 1, 1, '1598-01-01', 35750);
