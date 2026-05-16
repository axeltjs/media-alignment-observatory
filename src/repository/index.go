package repository

import "gorm.io/gorm"

type Registry struct {
	Source               *SourceRepository
	Article              *ArticleRepository
	GovernmentContent    *GovernmentContentRepository
	AnalysisResult       *AnalysisResultRepository
	CoordinationCluster  *CoordinationClusterRepository
}

func NewRegistry(db *gorm.DB) Registry {
	return Registry{
		Source:              &SourceRepository{DB: db},
		Article:             &ArticleRepository{DB: db},
		GovernmentContent:   &GovernmentContentRepository{DB: db},
		AnalysisResult:      &AnalysisResultRepository{DB: db},
		CoordinationCluster: &CoordinationClusterRepository{DB: db},
	}
}
