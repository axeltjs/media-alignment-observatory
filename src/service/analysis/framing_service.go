package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/hafidluqman50/maoi/src/model"
	"github.com/hafidluqman50/maoi/src/repository"
	"github.com/hafidluqman50/maoi/src/service/external"
)

const systemPromptFraming = `Kamu adalah analis media independen yang menganalisis narasi berita Indonesia secara objektif.
Tugasmu: bandingkan artikel media dengan pernyataan/konten resmi pemerintah, lalu hasilkan analisis framing.

Respond hanya dengan JSON valid, tanpa markdown, tanpa penjelasan tambahan.`

type FramingAnalysis struct {
	AlignmentScore  float64  `json:"alignment_score"`   // 0.0 - 1.0
	AlignmentLabel  string   `json:"alignment_label"`   // "high", "medium", "low"
	FramingType     string   `json:"framing_type"`      // "supportive", "neutral", "critical", "counter-narrative"
	KeyDifferences  []string `json:"key_differences"`
	KeySimilarities []string `json:"key_similarities"`
	Summary         string   `json:"summary"`
}

type FramingService struct {
	ArticleRepo        *repository.ArticleRepository
	GovContentRepo     *repository.GovernmentContentRepository
	AnalysisResultRepo *repository.AnalysisResultRepository
	ClaudeClient       *external.ClaudeClient
}

func (s *FramingService) IsConfigured() bool {
	return s.ClaudeClient != nil && s.ClaudeClient.IsConfigured()
}

// AnalyzeArticleFraming uses Claude to perform deep framing analysis between
// a media article and a government content piece.
func (s *FramingService) AnalyzeArticleFraming(ctx context.Context, article model.Article, govContent model.GovernmentContent) (*FramingAnalysis, error) {
	if !s.IsConfigured() {
		return nil, fmt.Errorf("framing service not configured")
	}

	content := article.CleanContent
	if content == "" {
		content = article.Content
	}
	govText := govContent.CleanContent
	if govText == "" {
		govText = govContent.Content
	}

	// Truncate to avoid token limits while keeping context meaningful.
	content = truncate(content, 2000)
	govText = truncate(govText, 2000)

	prompt := fmt.Sprintf(`Analisis perbandingan narasi berikut:

=== ARTIKEL MEDIA ===
Judul: %s
Konten: %s

=== KONTEN RESMI PEMERINTAH ===
Judul: %s
Sumber: %s
Konten: %s

Hasilkan analisis dalam format JSON:
{
  "alignment_score": <angka 0.0-1.0, di mana 1.0 = sepenuhnya selaras>,
  "alignment_label": <"high" jika >=0.7, "medium" jika >=0.4, "low" jika <0.4>,
  "framing_type": <"supportive"|"neutral"|"critical"|"counter-narrative">,
  "key_differences": [<list string perbedaan narasi kunci>],
  "key_similarities": [<list string kesamaan narasi kunci>],
  "summary": <ringkasan singkat 1-2 kalimat tentang alignment ini>
}`,
		article.Title, content,
		govContent.Title, govContent.Agency, govText,
	)

	raw, err := s.ClaudeClient.Complete(ctx, systemPromptFraming, prompt, 1024)
	if err != nil {
		return nil, fmt.Errorf("claude complete: %w", err)
	}

	raw = strings.TrimSpace(raw)
	var analysis FramingAnalysis
	if err := json.Unmarshal([]byte(raw), &analysis); err != nil {
		return nil, fmt.Errorf("parse framing response: %w (raw: %s)", err, raw)
	}
	return &analysis, nil
}

// AnalyzeAndSaveRecentArticles runs LLM framing analysis for recent articles
// that don't yet have a high-fidelity analysis result.
func (s *FramingService) AnalyzeAndSaveRecentArticles(ctx context.Context, articleLimit int) error {
	if !s.IsConfigured() {
		slog.Warn("framing service not configured, skipping")
		return nil
	}

	govContents, err := s.GovContentRepo.FindWithEmbedding(ctx)
	if err != nil {
		return fmt.Errorf("fetch gov contents: %w", err)
	}
	if len(govContents) == 0 {
		return nil
	}

	articles, err := s.ArticleRepo.FindRecent(ctx, articleLimit)
	if err != nil {
		return fmt.Errorf("fetch articles: %w", err)
	}

	for _, article := range articles {
		for _, gc := range govContents {
			exists, _ := s.AnalysisResultRepo.ExistsByArticleAndGovContent(ctx, article.ID, gc.ID)
			if exists {
				continue
			}

			analysis, err := s.AnalyzeArticleFraming(ctx, article, gc)
			if err != nil {
				slog.Warn("framing analysis failed", "article", article.ID, "gov", gc.ID, "error", err)
				continue
			}

			_, err = s.AnalysisResultRepo.Create(ctx, repository.AnalysisResultCreateInput{
				ArticleID:           article.ID,
				GovernmentContentID: gc.ID,
				SimilarityScore:     analysis.AlignmentScore,
				AlignmentLabel:      analysis.AlignmentLabel,
			})
			if err != nil {
				slog.Warn("save framing result failed", "error", err)
			}
		}
	}
	return nil
}

// SummarizeCluster uses Claude to generate a human-readable topic summary
// for a detected coordination cluster.
func (s *FramingService) SummarizeCluster(ctx context.Context, articleTitles []string) (string, error) {
	if !s.IsConfigured() {
		return "", fmt.Errorf("framing service not configured")
	}

	prompt := fmt.Sprintf(`Berikut adalah judul-judul artikel dari berbagai media Indonesia yang terdeteksi memiliki narasi sangat mirip dalam waktu singkat:

%s

Dalam 1-2 kalimat, ringkas: apa topik utama kluster ini dan mengapa ini menarik dari perspektif pemantauan media?`,
		strings.Join(articleTitles, "\n- "),
	)

	return s.ClaudeClient.Complete(ctx, "Kamu adalah analis media independen Indonesia.", prompt, 256)
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
