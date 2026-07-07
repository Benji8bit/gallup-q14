export const SURVEY_DIRECTION = 'Data Engineering';

export const SURVEY_ROLES = [
  { value: 'Менеджер проекта', labelRu: 'Менеджер проекта' },
  { value: 'Тимлид', labelRu: 'Тимлид' },
  { value: 'Инженер данных/Разработчик', labelRu: 'Инженер данных/Разработчик' },
] as const;

export type SurveyRole = (typeof SURVEY_ROLES)[number]['value'];
