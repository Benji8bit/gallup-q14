-- User feedback (07.2026): 13 questions (E01 + Q01–Q12), drop Q00 from active survey.

UPDATE questions SET is_active = 0 WHERE id = 'Q00';

UPDATE questions SET sort_order = 0 WHERE id = 'E01';
UPDATE questions SET sort_order = 1 WHERE id = 'Q01';
UPDATE questions SET sort_order = 2 WHERE id = 'Q02';
UPDATE questions SET sort_order = 3 WHERE id = 'Q03';
UPDATE questions SET sort_order = 4 WHERE id = 'Q04';
UPDATE questions SET sort_order = 5 WHERE id = 'Q05';
UPDATE questions SET sort_order = 6 WHERE id = 'Q06';
UPDATE questions SET sort_order = 7 WHERE id = 'Q07';
UPDATE questions SET sort_order = 8 WHERE id = 'Q08';
UPDATE questions SET sort_order = 9 WHERE id = 'Q09';
UPDATE questions SET sort_order = 10 WHERE id = 'Q10';
UPDATE questions SET sort_order = 11 WHERE id = 'Q11';
UPDATE questions SET sort_order = 12 WHERE id = 'Q12';
