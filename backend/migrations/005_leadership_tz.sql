-- ТЗ руководства (А. Гельмут, 06.07.2026): Gallup Q12 DE + eNPS, шкала согласия 1–5

ALTER TABLE questions ADD COLUMN scale_min INTEGER NOT NULL DEFAULT 1;
ALTER TABLE questions ADD COLUMN scale_max INTEGER NOT NULL DEFAULT 5;
ALTER TABLE questions ADD COLUMN question_role TEXT NOT NULL DEFAULT 'engagement';

UPDATE questions SET
    text_ru = 'В целом, насколько вы удовлетворены работой в Sapiens Solutions?',
    scale_min = 1, scale_max = 5, question_role = 'satisfaction'
WHERE id = 'Q00';

UPDATE questions SET
    text_ru = 'Я чётко понимаю, что от меня ожидается на проектах для клиентов (зона ответственности, результаты, сроки).',
    dimension = 'Basic Needs', sort_order = 1, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q01';

UPDATE questions SET
    text_ru = 'У меня есть все необходимые доступы, данные, инструменты и инфраструктура, чтобы качественно выполнять задачи по дата-инжинирингу.',
    dimension = 'Basic Needs', sort_order = 2, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q02';

UPDATE questions SET
    text_ru = 'В большинстве задач я могу использовать свои сильные стороны (технологии, архитектура, аналитика, коммуникация с заказчиком и т.п.).',
    dimension = 'Basic Needs', sort_order = 3, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q03';

UPDATE questions SET
    text_ru = 'За последние 2 недели я получал(а) признание или конструктивную позитивную обратную связь за вклад в проект (от руководителя, лида или клиента).',
    dimension = 'Individual Contribution', sort_order = 4, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q04';

UPDATE questions SET
    text_ru = 'Мой руководитель/тимлид демонстрирует, что ему не всё равно, как я себя чувствую (нагрузка, баланс, развитие, состояние).',
    dimension = 'Individual Contribution', sort_order = 5, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q05';

UPDATE questions SET
    text_ru = 'В компании есть человек (руководитель, ментор, старший коллега), который реально помогает мне профессионально расти в сфере дата-инжиниринга.',
    dimension = 'Individual Contribution', sort_order = 6, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q06';

UPDATE questions SET
    text_ru = 'На проектах и внутри команды считаются с моим мнением по техническим и организационным решениям.',
    dimension = 'Teamwork', sort_order = 7, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q07';

UPDATE questions SET
    text_ru = 'Миссия и проекты компании (то, как мы работаем с данными клиентов) заставляют меня чувствовать, что моя работа имеет смысл и влияет на бизнес заказчиков.',
    dimension = 'Teamwork', sort_order = 8, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q08';

UPDATE questions SET
    text_ru = 'Мои коллеги по проектам и командам обычно стремятся делать работу качественно, а не «просто закрывать задачи по трекеру».',
    dimension = 'Teamwork', sort_order = 9, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q09';

UPDATE questions SET
    text_ru = 'У меня есть хотя бы один коллега в компании, с кем у меня доверительные и по-настоящему хорошие рабочие отношения.',
    dimension = 'Teamwork', sort_order = 10, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q10';

UPDATE questions SET
    text_ru = 'За последние 6 месяцев я обсуждал(а) с руководителем или ментором свои результаты, сильные стороны и зоны роста (не только зарплату и загрузку).',
    dimension = 'Growth', sort_order = 11, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q11';

UPDATE questions SET
    text_ru = 'За последний год у меня была возможность учиться и развиваться как дата-инженер/консультант (новые технологии, архитектурные решения, участие в сложных проектах, внутренние инициативы).',
    dimension = 'Growth', sort_order = 12, scale_min = 1, scale_max = 5, question_role = 'engagement', is_active = 1
WHERE id = 'Q12';

-- Снять с опроса устаревшие Q13/Q14 и S01/S02 (исторические ответы сохраняются)
UPDATE questions SET is_active = 0 WHERE id IN ('Q13', 'Q14', 'S01', 'S02');

INSERT INTO questions (id, code, text_ru, dimension, sort_order, is_active, scale_min, scale_max, question_role) VALUES
('E01', 'E01',
 'По шкале от 0 до 10, насколько вероятно, что вы порекомендуете нашу компанию как место работы другу или коллеге?',
 'eNPS', 13, 1, 0, 10, 'enps')
ON CONFLICT(id) DO UPDATE SET
    text_ru = excluded.text_ru,
    dimension = excluded.dimension,
    sort_order = excluded.sort_order,
    is_active = excluded.is_active,
    scale_min = excluded.scale_min,
    scale_max = excluded.scale_max,
    question_role = excluded.question_role;

-- Расширить допустимый диапазон ответов (0–10 для eNPS)
CREATE TABLE IF NOT EXISTS answers_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_id INTEGER NOT NULL,
    question_id TEXT NOT NULL,
    value INTEGER NOT NULL CHECK(value BETWEEN 0 AND 10),
    FOREIGN KEY(submission_id) REFERENCES submissions(id) ON DELETE CASCADE,
    FOREIGN KEY(question_id) REFERENCES questions(id),
    UNIQUE(submission_id, question_id)
);

INSERT INTO answers_new (id, submission_id, question_id, value)
SELECT id, submission_id, question_id, value FROM answers;

DROP TABLE answers;
ALTER TABLE answers_new RENAME TO answers;
