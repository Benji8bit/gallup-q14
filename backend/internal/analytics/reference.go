package analytics

import "github.com/sapiens-solutions/gallup-q14/internal/models"

func dimensionLabels() map[string]string {
	return map[string]string{
		"Basic Needs":             "Роль и ресурсы",
		"Individual Contribution": "Признание и поддержка",
		"Teamwork":                "Голос, смысл и команда",
		"Growth":                  "Обратная связь и развитие",
	}
}

func dimensionAudiences() map[string]string {
	return map[string]string{
		"Basic Needs":             "Руководители проектов, Team Lead",
		"Individual Contribution": "Линейные руководители, HR",
		"Teamwork":                "PMO, руководители практик",
		"Growth":                  "HR, руководители практик, менторы",
	}
}

func questionMeta() map[string]struct {
	whatItMeasures  string
	leadershipFocus string
} {
	return map[string]struct {
		whatItMeasures  string
		leadershipFocus string
	}{
		"Q00": {"Общая удовлетворённость компанией", "Отдельный индикатор настроения, не входит в % вовлечённости"},
		"Q01": {"Ясность ожиданий на клиентских проектах", "Проверьте описания ролей, KPI, зоны ответственности и сроки"},
		"Q02": {"Доступы, данные, инструменты DE", "Аудит доступов, лицензий, CI/CD, сред разработки и инфраструктуры"},
		"Q03": {"Использование сильных сторон", "Подбор задач под компетенции: технологии, архитектура, аналитика, коммуникация"},
		"Q04": {"Признание за последние 2 недели", "Норма регулярной позитивной обратной связи от лида, руководителя или клиента"},
		"Q05": {"Забота руководства о человеке", "Качество 1:1, внимание к нагрузке, балансу и состоянию"},
		"Q06": {"Профессиональный рост в DE", "Наставничество, карьерные разговоры, доступ к экспертизе старших коллег"},
		"Q07": {"Учёт мнения по тех. и орг. решениям", "Вовлечение в решения, прозрачность «почему так»"},
		"Q08": {"Смысл работы и влияние на бизнес клиентов", "Связь задач с ценностью для заказчика и миссией компании"},
		"Q09": {"Культура качества в команде", "Стандарты delivery, code review, ответственность за результат"},
		"Q10": {"Доверительные рабочие отношения", "Командные ритуалы, онбординг, неформальное общение"},
		"Q11": {"Разговоры о прогрессе за 6 месяцев", "Регулярные career/check-in сессии не только про зарплату и загрузку"},
		"Q12": {"Обучение и развитие за год", "Бюджет обучения, сложные проекты, внутренние инициативы, guilds"},
		"E01": {"Готовность рекомендовать компанию (eNPS)", "Отдельный индекс лояльности; сегментируйте промоутеров, нейтралов и критиков"},
	}
}

func BuildMethodologyGuide(questionTexts map[string]string, questionScores []models.QuestionScore) models.MethodologyGuide {
	scoreByCode := map[string]models.QuestionScore{}
	for _, q := range questionScores {
		scoreByCode[q.Code] = q
	}

	meta := questionMeta()
	questions := make([]models.QuestionGuideItem, 0, len(questionTexts))
	codes := []string{"Q00", "Q01", "Q02", "Q03", "Q04", "Q05", "Q06", "Q07", "Q08", "Q09", "Q10", "Q11", "Q12", "E01"}
	for _, code := range codes {
		text := questionTexts[code]
		if text == "" {
			continue
		}
		m := meta[code]
		item := models.QuestionGuideItem{
			Code:            code,
			TextRU:          text,
			WhatItMeasures:  m.whatItMeasures,
			LeadershipFocus: m.leadershipFocus,
		}
		if sc, ok := scoreByCode[code]; ok {
			item.Dimension = sc.Dimension
			item.CurrentAvg5pt = sc.AverageScore5pt
			item.CurrentFavPct = sc.FavorablePct
		}
		questions = append(questions, item)
	}

	dimensions := make([]models.DimensionGuideItem, 0, len(dimensionLabels()))
	for key, label := range dimensionLabels() {
		dimensions = append(dimensions, models.DimensionGuideItem{
			Key:             key,
			LabelRU:         label,
			Description:     dimensionDescriptions()[key],
			PrimaryAudience: dimensionAudiences()[key],
		})
	}

	return models.MethodologyGuide{
		Overview: "Опрос вовлечённости Sapiens Solutions по ТЗ руководства (06.07.2026): адаптированный Gallup Q12 для data engineering консалтинга " +
			"(Q01–Q12, шкала согласия 1–5) + Q00 (удовлетворённость) + E01 (eNPS 0–10). " +
			"Индекс вовлечённости — доля благоприятных ответов (4–5) по Q01–Q12. eNPS рассчитывается отдельно.",
		Metrics: []models.MetricGuideItem{
			{
				Name:        "Индекс вовлечённости",
				Description: "Доля ответов «благоприятно» (4–5) по Q01–Q12.",
				Formula:     "favorable / все ответы Q01–Q12 × 100%",
				Scale:       "0–100%",
			},
			{
				Name:        "Удовлетворённость (Q00)",
				Description: "Средняя оценка удовлетворённости компанией; отдельный показатель.",
				Formula:     "среднее(Q00) на шкале 1–5",
				Scale:       "1–5",
			},
			{
				Name:        "eNPS (E01)",
				Description: "Индекс готовности рекомендовать компанию как место работы.",
				Formula:     "% промоутеров (9–10) − % критиков (0–6); нейтралы 7–8 не входят в формулу",
				Scale:       "от −100 до +100; >+30 хорошо, <0 тревожно",
			},
			{
				Name:        "% favorable по вопросу",
				Description: "Доля ответов 4–5 (согласен / полностью согласен) по Q01–Q12.",
				Formula:     "ответы ≥ 4 / все ответы × 100%",
				Scale:       "0–100%",
			},
		},
		Scales: []models.ScaleGuideSection{
			{
				Title:       "Q00 — удовлетворённость",
				Description: "Шкала 1–5: от «крайне не удовлетворён» до «полностью удовлетворён».",
			},
			{
				Title:       "Q01–Q12 — согласие (ТЗ руководства)",
				Description: "Шкала 1–5: 1 — «совершенно не согласен», 5 — «полностью согласен». Благоприятный ответ: 4–5.",
				Mapping: []models.ScaleMappingRow{
					{SurveyValue: "1–2", DashboardValue: "1–2", Meaning: "Негативная зона"},
					{SurveyValue: "3", DashboardValue: "3", Meaning: "Нейтральная зона"},
					{SurveyValue: "4–5", DashboardValue: "4–5", Meaning: "Favorable — согласен / полностью согласен"},
				},
			},
			{
				Title:       "E01 — eNPS",
				Description: "Шкала 0–10. Промоутеры 9–10, нейтралы 7–8, критики 0–6.",
				Mapping: []models.ScaleMappingRow{
					{SurveyValue: "0–6", DashboardValue: "Критики", Meaning: "Снижают eNPS"},
					{SurveyValue: "7–8", DashboardValue: "Нейтралы", Meaning: "Не входят в формулу eNPS"},
					{SurveyValue: "9–10", DashboardValue: "Промоутеры", Meaning: "Повышают eNPS"},
				},
			},
		},
		Questions:  questions,
		Dimensions: dimensions,
	}
}

func dimensionDescriptions() map[string]string {
	return map[string]string{
		"Basic Needs":             "Ясность ожиданий на проектах, доступы и инструменты DE, использование сильных сторон.",
		"Individual Contribution": "Признание, забота руководства, менторство и профессиональный рост.",
		"Teamwork":                "Учёт мнения, смысл работы, культура качества и доверительные связи в команде.",
		"Growth":                  "Обратная связь о прогрессе и возможности учиться как DE-консультант.",
	}
}
