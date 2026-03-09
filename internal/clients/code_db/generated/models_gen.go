package generated

// SortDirection represents sort ordering.
type SortDirection string

const (
	SortDirectionASC  SortDirection = "ASC"
	SortDirectionDESC SortDirection = "DESC"
)

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
}

type DeleteInfo struct {
	NodesDeleted int `json:"nodesDeleted"`
}

type CallProperties struct {
	CallType *string `json:"callType,omitempty"`
}

type CallPropertiesCreateInput struct {
	CallType *string `json:"callType,omitempty"`
}

type CallPropertiesUpdateInput struct {
	CallType *string `json:"callType,omitempty"`
}

type Class struct {
	Id            string        `json:"id"`
	Name          string        `json:"name"`
	Path          string        `json:"path"`
	Language      *string       `json:"language,omitempty"`
	Kind          string        `json:"kind"`
	Visibility    *string       `json:"visibility,omitempty"`
	Source        *string       `json:"source,omitempty"`
	StartingLine  *int          `json:"startingLine,omitempty"`
	EndingLine    *int          `json:"endingLine,omitempty"`
	Decorators    []string      `json:"decorators,omitempty"`
	Methods       []*Function   `json:"methods,omitempty"`
	Inherits      []*Class      `json:"inherits,omitempty"`
	InheritedBy   []*Class      `json:"inheritedBy,omitempty"`
	Implements    []*Class      `json:"implements,omitempty"`
	ImplementedBy []*Class      `json:"implementedBy,omitempty"`
	DefinedIn     []*File       `json:"definedIn,omitempty"`
	Module        []*Module     `json:"module,omitempty"`
	Repository    []*Repository `json:"repository,omitempty"`
}

type ClassCreateInput struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Language     *string  `json:"language,omitempty"`
	Kind         string   `json:"kind"`
	Visibility   *string  `json:"visibility,omitempty"`
	Source       *string  `json:"source,omitempty"`
	StartingLine *int     `json:"startingLine,omitempty"`
	EndingLine   *int     `json:"endingLine,omitempty"`
	Decorators   []string `json:"decorators,omitempty"`
}

type ClassUpdateInput struct {
	Name         *string  `json:"name,omitempty"`
	Path         *string  `json:"path,omitempty"`
	Language     *string  `json:"language,omitempty"`
	Kind         *string  `json:"kind,omitempty"`
	Visibility   *string  `json:"visibility,omitempty"`
	Source       *string  `json:"source,omitempty"`
	StartingLine *int     `json:"startingLine,omitempty"`
	EndingLine   *int     `json:"endingLine,omitempty"`
	Decorators   []string `json:"decorators,omitempty"`
}

type ClassWhere struct {
	Id                   *string       `json:"id,omitempty"`
	IdNot                *string       `json:"id_NOT,omitempty"`
	IdIn                 []string      `json:"id_IN,omitempty"`
	IdNotIn              []string      `json:"id_NOT_IN,omitempty"`
	IdContains           *string       `json:"id_CONTAINS,omitempty"`
	IdStartsWith         *string       `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith           *string       `json:"id_ENDS_WITH,omitempty"`
	Name                 *string       `json:"name,omitempty"`
	NameNot              *string       `json:"name_NOT,omitempty"`
	NameIn               []string      `json:"name_IN,omitempty"`
	NameNotIn            []string      `json:"name_NOT_IN,omitempty"`
	NameGt               *string       `json:"name_GT,omitempty"`
	NameGte              *string       `json:"name_GTE,omitempty"`
	NameLt               *string       `json:"name_LT,omitempty"`
	NameLte              *string       `json:"name_LTE,omitempty"`
	NameContains         *string       `json:"name_CONTAINS,omitempty"`
	NameStartsWith       *string       `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith         *string       `json:"name_ENDS_WITH,omitempty"`
	Path                 *string       `json:"path,omitempty"`
	PathNot              *string       `json:"path_NOT,omitempty"`
	PathIn               []string      `json:"path_IN,omitempty"`
	PathNotIn            []string      `json:"path_NOT_IN,omitempty"`
	PathGt               *string       `json:"path_GT,omitempty"`
	PathGte              *string       `json:"path_GTE,omitempty"`
	PathLt               *string       `json:"path_LT,omitempty"`
	PathLte              *string       `json:"path_LTE,omitempty"`
	PathContains         *string       `json:"path_CONTAINS,omitempty"`
	PathStartsWith       *string       `json:"path_STARTS_WITH,omitempty"`
	PathEndsWith         *string       `json:"path_ENDS_WITH,omitempty"`
	Language             *string       `json:"language,omitempty"`
	LanguageNot          *string       `json:"language_NOT,omitempty"`
	LanguageIn           []string      `json:"language_IN,omitempty"`
	LanguageNotIn        []string      `json:"language_NOT_IN,omitempty"`
	LanguageGt           *string       `json:"language_GT,omitempty"`
	LanguageGte          *string       `json:"language_GTE,omitempty"`
	LanguageLt           *string       `json:"language_LT,omitempty"`
	LanguageLte          *string       `json:"language_LTE,omitempty"`
	LanguageContains     *string       `json:"language_CONTAINS,omitempty"`
	LanguageStartsWith   *string       `json:"language_STARTS_WITH,omitempty"`
	LanguageEndsWith     *string       `json:"language_ENDS_WITH,omitempty"`
	Kind                 *string       `json:"kind,omitempty"`
	KindNot              *string       `json:"kind_NOT,omitempty"`
	KindIn               []string      `json:"kind_IN,omitempty"`
	KindNotIn            []string      `json:"kind_NOT_IN,omitempty"`
	KindGt               *string       `json:"kind_GT,omitempty"`
	KindGte              *string       `json:"kind_GTE,omitempty"`
	KindLt               *string       `json:"kind_LT,omitempty"`
	KindLte              *string       `json:"kind_LTE,omitempty"`
	KindContains         *string       `json:"kind_CONTAINS,omitempty"`
	KindStartsWith       *string       `json:"kind_STARTS_WITH,omitempty"`
	KindEndsWith         *string       `json:"kind_ENDS_WITH,omitempty"`
	Visibility           *string       `json:"visibility,omitempty"`
	VisibilityNot        *string       `json:"visibility_NOT,omitempty"`
	VisibilityIn         []string      `json:"visibility_IN,omitempty"`
	VisibilityNotIn      []string      `json:"visibility_NOT_IN,omitempty"`
	VisibilityGt         *string       `json:"visibility_GT,omitempty"`
	VisibilityGte        *string       `json:"visibility_GTE,omitempty"`
	VisibilityLt         *string       `json:"visibility_LT,omitempty"`
	VisibilityLte        *string       `json:"visibility_LTE,omitempty"`
	VisibilityContains   *string       `json:"visibility_CONTAINS,omitempty"`
	VisibilityStartsWith *string       `json:"visibility_STARTS_WITH,omitempty"`
	VisibilityEndsWith   *string       `json:"visibility_ENDS_WITH,omitempty"`
	Source               *string       `json:"source,omitempty"`
	SourceNot            *string       `json:"source_NOT,omitempty"`
	SourceIn             []string      `json:"source_IN,omitempty"`
	SourceNotIn          []string      `json:"source_NOT_IN,omitempty"`
	SourceGt             *string       `json:"source_GT,omitempty"`
	SourceGte            *string       `json:"source_GTE,omitempty"`
	SourceLt             *string       `json:"source_LT,omitempty"`
	SourceLte            *string       `json:"source_LTE,omitempty"`
	SourceContains       *string       `json:"source_CONTAINS,omitempty"`
	SourceStartsWith     *string       `json:"source_STARTS_WITH,omitempty"`
	SourceEndsWith       *string       `json:"source_ENDS_WITH,omitempty"`
	StartingLine         *int          `json:"startingLine,omitempty"`
	StartingLineNot      *int          `json:"startingLine_NOT,omitempty"`
	StartingLineIn       []int         `json:"startingLine_IN,omitempty"`
	StartingLineNotIn    []int         `json:"startingLine_NOT_IN,omitempty"`
	StartingLineGt       *int          `json:"startingLine_GT,omitempty"`
	StartingLineGte      *int          `json:"startingLine_GTE,omitempty"`
	StartingLineLt       *int          `json:"startingLine_LT,omitempty"`
	StartingLineLte      *int          `json:"startingLine_LTE,omitempty"`
	EndingLine           *int          `json:"endingLine,omitempty"`
	EndingLineNot        *int          `json:"endingLine_NOT,omitempty"`
	EndingLineIn         []int         `json:"endingLine_IN,omitempty"`
	EndingLineNotIn      []int         `json:"endingLine_NOT_IN,omitempty"`
	EndingLineGt         *int          `json:"endingLine_GT,omitempty"`
	EndingLineGte        *int          `json:"endingLine_GTE,omitempty"`
	EndingLineLt         *int          `json:"endingLine_LT,omitempty"`
	EndingLineLte        *int          `json:"endingLine_LTE,omitempty"`
	Decorators           *[]string     `json:"decorators,omitempty"`
	DecoratorsNot        *[]string     `json:"decorators_NOT,omitempty"`
	DecoratorsIn         [][]string    `json:"decorators_IN,omitempty"`
	DecoratorsNotIn      [][]string    `json:"decorators_NOT_IN,omitempty"`
	DecoratorsGt         *[]string     `json:"decorators_GT,omitempty"`
	DecoratorsGte        *[]string     `json:"decorators_GTE,omitempty"`
	DecoratorsLt         *[]string     `json:"decorators_LT,omitempty"`
	DecoratorsLte        *[]string     `json:"decorators_LTE,omitempty"`
	AND                  []*ClassWhere `json:"AND,omitempty"`
	OR                   []*ClassWhere `json:"OR,omitempty"`
	NOT                  *ClassWhere   `json:"NOT,omitempty"`
}

type ClassSort struct {
	Id           *SortDirection `json:"id,omitempty"`
	Name         *SortDirection `json:"name,omitempty"`
	Path         *SortDirection `json:"path,omitempty"`
	Language     *SortDirection `json:"language,omitempty"`
	Kind         *SortDirection `json:"kind,omitempty"`
	Visibility   *SortDirection `json:"visibility,omitempty"`
	Source       *SortDirection `json:"source,omitempty"`
	StartingLine *SortDirection `json:"startingLine,omitempty"`
	EndingLine   *SortDirection `json:"endingLine,omitempty"`
	Decorators   *SortDirection `json:"decorators,omitempty"`
}

type ClasssConnection struct {
	Edges      []*ClassEdge `json:"edges"`
	TotalCount int          `json:"totalCount"`
	PageInfo   PageInfo     `json:"pageInfo"`
}

type ClassEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type CreateClasssMutationResponse struct {
	Classs []*Class `json:"classs"`
}

type UpdateClasssMutationResponse struct {
	Classs []*Class `json:"classs"`
}

type ClassMethodsConnection struct {
	Edges      []*ClassMethodsEdge `json:"edges"`
	TotalCount int                 `json:"totalCount"`
	PageInfo   PageInfo            `json:"pageInfo"`
}

type ClassMethodsEdge struct {
	Node   *Function `json:"node"`
	Cursor string    `json:"cursor"`
}

type ClassMethodsFieldInput struct {
	Create  []*ClassMethodsCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassMethodsConnectFieldInput `json:"connect,omitempty"`
}

type ClassMethodsUpdateFieldInput struct {
	Create     []*ClassMethodsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassMethodsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassMethodsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassMethodsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassMethodsDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassMethodsCreateFieldInput struct {
	Node FunctionCreateInput `json:"node"`
}

type ClassMethodsConnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type ClassMethodsDisconnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type ClassMethodsUpdateConnectionInput struct {
	Where FunctionWhere        `json:"where"`
	Node  *FunctionUpdateInput `json:"node,omitempty"`
}

type ClassMethodsDeleteFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type ClassInheritsConnection struct {
	Edges      []*ClassInheritsEdge `json:"edges"`
	TotalCount int                  `json:"totalCount"`
	PageInfo   PageInfo             `json:"pageInfo"`
}

type ClassInheritsEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type ClassInheritsFieldInput struct {
	Create  []*ClassInheritsCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassInheritsConnectFieldInput `json:"connect,omitempty"`
}

type ClassInheritsUpdateFieldInput struct {
	Create     []*ClassInheritsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassInheritsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassInheritsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassInheritsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassInheritsDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassInheritsCreateFieldInput struct {
	Node ClassCreateInput `json:"node"`
}

type ClassInheritsConnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassInheritsDisconnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassInheritsUpdateConnectionInput struct {
	Where ClassWhere        `json:"where"`
	Node  *ClassUpdateInput `json:"node,omitempty"`
}

type ClassInheritsDeleteFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassInheritedByConnection struct {
	Edges      []*ClassInheritedByEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type ClassInheritedByEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type ClassInheritedByFieldInput struct {
	Create  []*ClassInheritedByCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassInheritedByConnectFieldInput `json:"connect,omitempty"`
}

type ClassInheritedByUpdateFieldInput struct {
	Create     []*ClassInheritedByCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassInheritedByConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassInheritedByDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassInheritedByUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassInheritedByDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassInheritedByCreateFieldInput struct {
	Node ClassCreateInput `json:"node"`
}

type ClassInheritedByConnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassInheritedByDisconnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassInheritedByUpdateConnectionInput struct {
	Where ClassWhere        `json:"where"`
	Node  *ClassUpdateInput `json:"node,omitempty"`
}

type ClassInheritedByDeleteFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassImplementsConnection struct {
	Edges      []*ClassImplementsEdge `json:"edges"`
	TotalCount int                    `json:"totalCount"`
	PageInfo   PageInfo               `json:"pageInfo"`
}

type ClassImplementsEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type ClassImplementsFieldInput struct {
	Create  []*ClassImplementsCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassImplementsConnectFieldInput `json:"connect,omitempty"`
}

type ClassImplementsUpdateFieldInput struct {
	Create     []*ClassImplementsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassImplementsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassImplementsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassImplementsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassImplementsDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassImplementsCreateFieldInput struct {
	Node ClassCreateInput `json:"node"`
}

type ClassImplementsConnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassImplementsDisconnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassImplementsUpdateConnectionInput struct {
	Where ClassWhere        `json:"where"`
	Node  *ClassUpdateInput `json:"node,omitempty"`
}

type ClassImplementsDeleteFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassImplementedByConnection struct {
	Edges      []*ClassImplementedByEdge `json:"edges"`
	TotalCount int                       `json:"totalCount"`
	PageInfo   PageInfo                  `json:"pageInfo"`
}

type ClassImplementedByEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type ClassImplementedByFieldInput struct {
	Create  []*ClassImplementedByCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassImplementedByConnectFieldInput `json:"connect,omitempty"`
}

type ClassImplementedByUpdateFieldInput struct {
	Create     []*ClassImplementedByCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassImplementedByConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassImplementedByDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassImplementedByUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassImplementedByDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassImplementedByCreateFieldInput struct {
	Node ClassCreateInput `json:"node"`
}

type ClassImplementedByConnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassImplementedByDisconnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassImplementedByUpdateConnectionInput struct {
	Where ClassWhere        `json:"where"`
	Node  *ClassUpdateInput `json:"node,omitempty"`
}

type ClassImplementedByDeleteFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ClassDefinedInConnection struct {
	Edges      []*ClassDefinedInEdge `json:"edges"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   PageInfo              `json:"pageInfo"`
}

type ClassDefinedInEdge struct {
	Node   *File  `json:"node"`
	Cursor string `json:"cursor"`
}

type ClassDefinedInFieldInput struct {
	Create  []*ClassDefinedInCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassDefinedInConnectFieldInput `json:"connect,omitempty"`
}

type ClassDefinedInUpdateFieldInput struct {
	Create     []*ClassDefinedInCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassDefinedInConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassDefinedInDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassDefinedInUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassDefinedInDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassDefinedInCreateFieldInput struct {
	Node FileCreateInput `json:"node"`
}

type ClassDefinedInConnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type ClassDefinedInDisconnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type ClassDefinedInUpdateConnectionInput struct {
	Where FileWhere        `json:"where"`
	Node  *FileUpdateInput `json:"node,omitempty"`
}

type ClassDefinedInDeleteFieldInput struct {
	Where FileWhere `json:"where"`
}

type ClassModuleConnection struct {
	Edges      []*ClassModuleEdge `json:"edges"`
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
}

type ClassModuleEdge struct {
	Node   *Module `json:"node"`
	Cursor string  `json:"cursor"`
}

type ClassModuleFieldInput struct {
	Create  []*ClassModuleCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassModuleConnectFieldInput `json:"connect,omitempty"`
}

type ClassModuleUpdateFieldInput struct {
	Create     []*ClassModuleCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassModuleConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassModuleDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassModuleUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassModuleDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassModuleCreateFieldInput struct {
	Node ModuleCreateInput `json:"node"`
}

type ClassModuleConnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ClassModuleDisconnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ClassModuleUpdateConnectionInput struct {
	Where ModuleWhere        `json:"where"`
	Node  *ModuleUpdateInput `json:"node,omitempty"`
}

type ClassModuleDeleteFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ClassRepositoryConnection struct {
	Edges      []*ClassRepositoryEdge `json:"edges"`
	TotalCount int                    `json:"totalCount"`
	PageInfo   PageInfo               `json:"pageInfo"`
}

type ClassRepositoryEdge struct {
	Node   *Repository `json:"node"`
	Cursor string      `json:"cursor"`
}

type ClassRepositoryFieldInput struct {
	Create  []*ClassRepositoryCreateFieldInput  `json:"create,omitempty"`
	Connect []*ClassRepositoryConnectFieldInput `json:"connect,omitempty"`
}

type ClassRepositoryUpdateFieldInput struct {
	Create     []*ClassRepositoryCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ClassRepositoryConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ClassRepositoryDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ClassRepositoryUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ClassRepositoryDeleteFieldInput      `json:"delete,omitempty"`
}

type ClassRepositoryCreateFieldInput struct {
	Node RepositoryCreateInput `json:"node"`
}

type ClassRepositoryConnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type ClassRepositoryDisconnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type ClassRepositoryUpdateConnectionInput struct {
	Where RepositoryWhere        `json:"where"`
	Node  *RepositoryUpdateInput `json:"node,omitempty"`
}

type ClassRepositoryDeleteFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type Folder struct {
	Id           string        `json:"id"`
	Path         string        `json:"path"`
	LastUpdated  DateTime      `json:"lastUpdated"`
	Subfolders   []*Folder     `json:"subfolders,omitempty"`
	Files        []*File       `json:"files,omitempty"`
	ParentFolder []*Folder     `json:"parentFolder,omitempty"`
	Repository   []*Repository `json:"repository,omitempty"`
}

type FolderCreateInput struct {
	Path        string   `json:"path"`
	LastUpdated DateTime `json:"lastUpdated"`
}

type FolderUpdateInput struct {
	Path        *string   `json:"path,omitempty"`
	LastUpdated *DateTime `json:"lastUpdated,omitempty"`
}

type FolderWhere struct {
	Id               *string        `json:"id,omitempty"`
	IdNot            *string        `json:"id_NOT,omitempty"`
	IdIn             []string       `json:"id_IN,omitempty"`
	IdNotIn          []string       `json:"id_NOT_IN,omitempty"`
	IdContains       *string        `json:"id_CONTAINS,omitempty"`
	IdStartsWith     *string        `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith       *string        `json:"id_ENDS_WITH,omitempty"`
	Path             *string        `json:"path,omitempty"`
	PathNot          *string        `json:"path_NOT,omitempty"`
	PathIn           []string       `json:"path_IN,omitempty"`
	PathNotIn        []string       `json:"path_NOT_IN,omitempty"`
	PathGt           *string        `json:"path_GT,omitempty"`
	PathGte          *string        `json:"path_GTE,omitempty"`
	PathLt           *string        `json:"path_LT,omitempty"`
	PathLte          *string        `json:"path_LTE,omitempty"`
	PathContains     *string        `json:"path_CONTAINS,omitempty"`
	PathStartsWith   *string        `json:"path_STARTS_WITH,omitempty"`
	PathEndsWith     *string        `json:"path_ENDS_WITH,omitempty"`
	LastUpdated      *DateTime      `json:"lastUpdated,omitempty"`
	LastUpdatedNot   *DateTime      `json:"lastUpdated_NOT,omitempty"`
	LastUpdatedIn    []DateTime     `json:"lastUpdated_IN,omitempty"`
	LastUpdatedNotIn []DateTime     `json:"lastUpdated_NOT_IN,omitempty"`
	LastUpdatedGt    *DateTime      `json:"lastUpdated_GT,omitempty"`
	LastUpdatedGte   *DateTime      `json:"lastUpdated_GTE,omitempty"`
	LastUpdatedLt    *DateTime      `json:"lastUpdated_LT,omitempty"`
	LastUpdatedLte   *DateTime      `json:"lastUpdated_LTE,omitempty"`
	AND              []*FolderWhere `json:"AND,omitempty"`
	OR               []*FolderWhere `json:"OR,omitempty"`
	NOT              *FolderWhere   `json:"NOT,omitempty"`
}

type FolderSort struct {
	Id          *SortDirection `json:"id,omitempty"`
	Path        *SortDirection `json:"path,omitempty"`
	LastUpdated *SortDirection `json:"lastUpdated,omitempty"`
}

type FoldersConnection struct {
	Edges      []*FolderEdge `json:"edges"`
	TotalCount int           `json:"totalCount"`
	PageInfo   PageInfo      `json:"pageInfo"`
}

type FolderEdge struct {
	Node   *Folder `json:"node"`
	Cursor string  `json:"cursor"`
}

type CreateFoldersMutationResponse struct {
	Folders []*Folder `json:"folders"`
}

type UpdateFoldersMutationResponse struct {
	Folders []*Folder `json:"folders"`
}

type FolderSubfoldersConnection struct {
	Edges      []*FolderSubfoldersEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type FolderSubfoldersEdge struct {
	Node   *Folder `json:"node"`
	Cursor string  `json:"cursor"`
}

type FolderSubfoldersFieldInput struct {
	Create  []*FolderSubfoldersCreateFieldInput  `json:"create,omitempty"`
	Connect []*FolderSubfoldersConnectFieldInput `json:"connect,omitempty"`
}

type FolderSubfoldersUpdateFieldInput struct {
	Create     []*FolderSubfoldersCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FolderSubfoldersConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FolderSubfoldersDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FolderSubfoldersUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FolderSubfoldersDeleteFieldInput      `json:"delete,omitempty"`
}

type FolderSubfoldersCreateFieldInput struct {
	Node FolderCreateInput `json:"node"`
}

type FolderSubfoldersConnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FolderSubfoldersDisconnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FolderSubfoldersUpdateConnectionInput struct {
	Where FolderWhere        `json:"where"`
	Node  *FolderUpdateInput `json:"node,omitempty"`
}

type FolderSubfoldersDeleteFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FolderFilesConnection struct {
	Edges      []*FolderFilesEdge `json:"edges"`
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
}

type FolderFilesEdge struct {
	Node   *File  `json:"node"`
	Cursor string `json:"cursor"`
}

type FolderFilesFieldInput struct {
	Create  []*FolderFilesCreateFieldInput  `json:"create,omitempty"`
	Connect []*FolderFilesConnectFieldInput `json:"connect,omitempty"`
}

type FolderFilesUpdateFieldInput struct {
	Create     []*FolderFilesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FolderFilesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FolderFilesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FolderFilesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FolderFilesDeleteFieldInput      `json:"delete,omitempty"`
}

type FolderFilesCreateFieldInput struct {
	Node FileCreateInput `json:"node"`
}

type FolderFilesConnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type FolderFilesDisconnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type FolderFilesUpdateConnectionInput struct {
	Where FileWhere        `json:"where"`
	Node  *FileUpdateInput `json:"node,omitempty"`
}

type FolderFilesDeleteFieldInput struct {
	Where FileWhere `json:"where"`
}

type FolderParentFolderConnection struct {
	Edges      []*FolderParentFolderEdge `json:"edges"`
	TotalCount int                       `json:"totalCount"`
	PageInfo   PageInfo                  `json:"pageInfo"`
}

type FolderParentFolderEdge struct {
	Node   *Folder `json:"node"`
	Cursor string  `json:"cursor"`
}

type FolderParentFolderFieldInput struct {
	Create  []*FolderParentFolderCreateFieldInput  `json:"create,omitempty"`
	Connect []*FolderParentFolderConnectFieldInput `json:"connect,omitempty"`
}

type FolderParentFolderUpdateFieldInput struct {
	Create     []*FolderParentFolderCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FolderParentFolderConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FolderParentFolderDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FolderParentFolderUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FolderParentFolderDeleteFieldInput      `json:"delete,omitempty"`
}

type FolderParentFolderCreateFieldInput struct {
	Node FolderCreateInput `json:"node"`
}

type FolderParentFolderConnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FolderParentFolderDisconnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FolderParentFolderUpdateConnectionInput struct {
	Where FolderWhere        `json:"where"`
	Node  *FolderUpdateInput `json:"node,omitempty"`
}

type FolderParentFolderDeleteFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FolderRepositoryConnection struct {
	Edges      []*FolderRepositoryEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type FolderRepositoryEdge struct {
	Node   *Repository `json:"node"`
	Cursor string      `json:"cursor"`
}

type FolderRepositoryFieldInput struct {
	Create  []*FolderRepositoryCreateFieldInput  `json:"create,omitempty"`
	Connect []*FolderRepositoryConnectFieldInput `json:"connect,omitempty"`
}

type FolderRepositoryUpdateFieldInput struct {
	Create     []*FolderRepositoryCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FolderRepositoryConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FolderRepositoryDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FolderRepositoryUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FolderRepositoryDeleteFieldInput      `json:"delete,omitempty"`
}

type FolderRepositoryCreateFieldInput struct {
	Node RepositoryCreateInput `json:"node"`
}

type FolderRepositoryConnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type FolderRepositoryDisconnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type FolderRepositoryUpdateConnectionInput struct {
	Where RepositoryWhere        `json:"where"`
	Node  *RepositoryUpdateInput `json:"node,omitempty"`
}

type FolderRepositoryDeleteFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type File struct {
	Id              string               `json:"id"`
	Path            string               `json:"path"`
	Filename        *string              `json:"filename,omitempty"`
	Language        *string              `json:"language,omitempty"`
	LineCount       *int                 `json:"lineCount,omitempty"`
	LastUpdated     DateTime             `json:"lastUpdated"`
	Functions       []*Function          `json:"functions,omitempty"`
	Classes         []*Class             `json:"classes,omitempty"`
	Imports         []*Module            `json:"imports,omitempty"`
	ExternalImports []*ExternalReference `json:"externalImports,omitempty"`
	Folder          []*Folder            `json:"folder,omitempty"`
	Repository      []*Repository        `json:"repository,omitempty"`
}

type FileCreateInput struct {
	Path        string   `json:"path"`
	Filename    *string  `json:"filename,omitempty"`
	Language    *string  `json:"language,omitempty"`
	LineCount   *int     `json:"lineCount,omitempty"`
	LastUpdated DateTime `json:"lastUpdated"`
}

type FileUpdateInput struct {
	Path        *string   `json:"path,omitempty"`
	Filename    *string   `json:"filename,omitempty"`
	Language    *string   `json:"language,omitempty"`
	LineCount   *int      `json:"lineCount,omitempty"`
	LastUpdated *DateTime `json:"lastUpdated,omitempty"`
}

type FileWhere struct {
	Id                 *string      `json:"id,omitempty"`
	IdNot              *string      `json:"id_NOT,omitempty"`
	IdIn               []string     `json:"id_IN,omitempty"`
	IdNotIn            []string     `json:"id_NOT_IN,omitempty"`
	IdContains         *string      `json:"id_CONTAINS,omitempty"`
	IdStartsWith       *string      `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith         *string      `json:"id_ENDS_WITH,omitempty"`
	Path               *string      `json:"path,omitempty"`
	PathNot            *string      `json:"path_NOT,omitempty"`
	PathIn             []string     `json:"path_IN,omitempty"`
	PathNotIn          []string     `json:"path_NOT_IN,omitempty"`
	PathGt             *string      `json:"path_GT,omitempty"`
	PathGte            *string      `json:"path_GTE,omitempty"`
	PathLt             *string      `json:"path_LT,omitempty"`
	PathLte            *string      `json:"path_LTE,omitempty"`
	PathContains       *string      `json:"path_CONTAINS,omitempty"`
	PathStartsWith     *string      `json:"path_STARTS_WITH,omitempty"`
	PathEndsWith       *string      `json:"path_ENDS_WITH,omitempty"`
	Filename           *string      `json:"filename,omitempty"`
	FilenameNot        *string      `json:"filename_NOT,omitempty"`
	FilenameIn         []string     `json:"filename_IN,omitempty"`
	FilenameNotIn      []string     `json:"filename_NOT_IN,omitempty"`
	FilenameGt         *string      `json:"filename_GT,omitempty"`
	FilenameGte        *string      `json:"filename_GTE,omitempty"`
	FilenameLt         *string      `json:"filename_LT,omitempty"`
	FilenameLte        *string      `json:"filename_LTE,omitempty"`
	FilenameContains   *string      `json:"filename_CONTAINS,omitempty"`
	FilenameStartsWith *string      `json:"filename_STARTS_WITH,omitempty"`
	FilenameEndsWith   *string      `json:"filename_ENDS_WITH,omitempty"`
	Language           *string      `json:"language,omitempty"`
	LanguageNot        *string      `json:"language_NOT,omitempty"`
	LanguageIn         []string     `json:"language_IN,omitempty"`
	LanguageNotIn      []string     `json:"language_NOT_IN,omitempty"`
	LanguageGt         *string      `json:"language_GT,omitempty"`
	LanguageGte        *string      `json:"language_GTE,omitempty"`
	LanguageLt         *string      `json:"language_LT,omitempty"`
	LanguageLte        *string      `json:"language_LTE,omitempty"`
	LanguageContains   *string      `json:"language_CONTAINS,omitempty"`
	LanguageStartsWith *string      `json:"language_STARTS_WITH,omitempty"`
	LanguageEndsWith   *string      `json:"language_ENDS_WITH,omitempty"`
	LineCount          *int         `json:"lineCount,omitempty"`
	LineCountNot       *int         `json:"lineCount_NOT,omitempty"`
	LineCountIn        []int        `json:"lineCount_IN,omitempty"`
	LineCountNotIn     []int        `json:"lineCount_NOT_IN,omitempty"`
	LineCountGt        *int         `json:"lineCount_GT,omitempty"`
	LineCountGte       *int         `json:"lineCount_GTE,omitempty"`
	LineCountLt        *int         `json:"lineCount_LT,omitempty"`
	LineCountLte       *int         `json:"lineCount_LTE,omitempty"`
	LastUpdated        *DateTime    `json:"lastUpdated,omitempty"`
	LastUpdatedNot     *DateTime    `json:"lastUpdated_NOT,omitempty"`
	LastUpdatedIn      []DateTime   `json:"lastUpdated_IN,omitempty"`
	LastUpdatedNotIn   []DateTime   `json:"lastUpdated_NOT_IN,omitempty"`
	LastUpdatedGt      *DateTime    `json:"lastUpdated_GT,omitempty"`
	LastUpdatedGte     *DateTime    `json:"lastUpdated_GTE,omitempty"`
	LastUpdatedLt      *DateTime    `json:"lastUpdated_LT,omitempty"`
	LastUpdatedLte     *DateTime    `json:"lastUpdated_LTE,omitempty"`
	AND                []*FileWhere `json:"AND,omitempty"`
	OR                 []*FileWhere `json:"OR,omitempty"`
	NOT                *FileWhere   `json:"NOT,omitempty"`
}

type FileSort struct {
	Id          *SortDirection `json:"id,omitempty"`
	Path        *SortDirection `json:"path,omitempty"`
	Filename    *SortDirection `json:"filename,omitempty"`
	Language    *SortDirection `json:"language,omitempty"`
	LineCount   *SortDirection `json:"lineCount,omitempty"`
	LastUpdated *SortDirection `json:"lastUpdated,omitempty"`
}

type FilesConnection struct {
	Edges      []*FileEdge `json:"edges"`
	TotalCount int         `json:"totalCount"`
	PageInfo   PageInfo    `json:"pageInfo"`
}

type FileEdge struct {
	Node   *File  `json:"node"`
	Cursor string `json:"cursor"`
}

type CreateFilesMutationResponse struct {
	Files []*File `json:"files"`
}

type UpdateFilesMutationResponse struct {
	Files []*File `json:"files"`
}

type FileFunctionsConnection struct {
	Edges      []*FileFunctionsEdge `json:"edges"`
	TotalCount int                  `json:"totalCount"`
	PageInfo   PageInfo             `json:"pageInfo"`
}

type FileFunctionsEdge struct {
	Node   *Function `json:"node"`
	Cursor string    `json:"cursor"`
}

type FileFunctionsFieldInput struct {
	Create  []*FileFunctionsCreateFieldInput  `json:"create,omitempty"`
	Connect []*FileFunctionsConnectFieldInput `json:"connect,omitempty"`
}

type FileFunctionsUpdateFieldInput struct {
	Create     []*FileFunctionsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FileFunctionsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FileFunctionsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FileFunctionsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FileFunctionsDeleteFieldInput      `json:"delete,omitempty"`
}

type FileFunctionsCreateFieldInput struct {
	Node FunctionCreateInput `json:"node"`
}

type FileFunctionsConnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FileFunctionsDisconnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FileFunctionsUpdateConnectionInput struct {
	Where FunctionWhere        `json:"where"`
	Node  *FunctionUpdateInput `json:"node,omitempty"`
}

type FileFunctionsDeleteFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FileClassesConnection struct {
	Edges      []*FileClassesEdge `json:"edges"`
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
}

type FileClassesEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type FileClassesFieldInput struct {
	Create  []*FileClassesCreateFieldInput  `json:"create,omitempty"`
	Connect []*FileClassesConnectFieldInput `json:"connect,omitempty"`
}

type FileClassesUpdateFieldInput struct {
	Create     []*FileClassesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FileClassesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FileClassesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FileClassesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FileClassesDeleteFieldInput      `json:"delete,omitempty"`
}

type FileClassesCreateFieldInput struct {
	Node ClassCreateInput `json:"node"`
}

type FileClassesConnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type FileClassesDisconnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type FileClassesUpdateConnectionInput struct {
	Where ClassWhere        `json:"where"`
	Node  *ClassUpdateInput `json:"node,omitempty"`
}

type FileClassesDeleteFieldInput struct {
	Where ClassWhere `json:"where"`
}

type FileImportsConnection struct {
	Edges      []*FileImportsEdge `json:"edges"`
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
}

type FileImportsEdge struct {
	Node   *Module `json:"node"`
	Cursor string  `json:"cursor"`
}

type FileImportsFieldInput struct {
	Create  []*FileImportsCreateFieldInput  `json:"create,omitempty"`
	Connect []*FileImportsConnectFieldInput `json:"connect,omitempty"`
}

type FileImportsUpdateFieldInput struct {
	Create     []*FileImportsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FileImportsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FileImportsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FileImportsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FileImportsDeleteFieldInput      `json:"delete,omitempty"`
}

type FileImportsCreateFieldInput struct {
	Node ModuleCreateInput `json:"node"`
}

type FileImportsConnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type FileImportsDisconnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type FileImportsUpdateConnectionInput struct {
	Where ModuleWhere        `json:"where"`
	Node  *ModuleUpdateInput `json:"node,omitempty"`
}

type FileImportsDeleteFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type FileExternalImportsConnection struct {
	Edges      []*FileExternalImportsEdge `json:"edges"`
	TotalCount int                        `json:"totalCount"`
	PageInfo   PageInfo                   `json:"pageInfo"`
}

type FileExternalImportsEdge struct {
	Node   *ExternalReference `json:"node"`
	Cursor string             `json:"cursor"`
}

type FileExternalImportsFieldInput struct {
	Create  []*FileExternalImportsCreateFieldInput  `json:"create,omitempty"`
	Connect []*FileExternalImportsConnectFieldInput `json:"connect,omitempty"`
}

type FileExternalImportsUpdateFieldInput struct {
	Create     []*FileExternalImportsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FileExternalImportsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FileExternalImportsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FileExternalImportsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FileExternalImportsDeleteFieldInput      `json:"delete,omitempty"`
}

type FileExternalImportsCreateFieldInput struct {
	Node ExternalReferenceCreateInput `json:"node"`
}

type FileExternalImportsConnectFieldInput struct {
	Where ExternalReferenceWhere `json:"where"`
}

type FileExternalImportsDisconnectFieldInput struct {
	Where ExternalReferenceWhere `json:"where"`
}

type FileExternalImportsUpdateConnectionInput struct {
	Where ExternalReferenceWhere        `json:"where"`
	Node  *ExternalReferenceUpdateInput `json:"node,omitempty"`
}

type FileExternalImportsDeleteFieldInput struct {
	Where ExternalReferenceWhere `json:"where"`
}

type FileFolderConnection struct {
	Edges      []*FileFolderEdge `json:"edges"`
	TotalCount int               `json:"totalCount"`
	PageInfo   PageInfo          `json:"pageInfo"`
}

type FileFolderEdge struct {
	Node   *Folder `json:"node"`
	Cursor string  `json:"cursor"`
}

type FileFolderFieldInput struct {
	Create  []*FileFolderCreateFieldInput  `json:"create,omitempty"`
	Connect []*FileFolderConnectFieldInput `json:"connect,omitempty"`
}

type FileFolderUpdateFieldInput struct {
	Create     []*FileFolderCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FileFolderConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FileFolderDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FileFolderUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FileFolderDeleteFieldInput      `json:"delete,omitempty"`
}

type FileFolderCreateFieldInput struct {
	Node FolderCreateInput `json:"node"`
}

type FileFolderConnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FileFolderDisconnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FileFolderUpdateConnectionInput struct {
	Where FolderWhere        `json:"where"`
	Node  *FolderUpdateInput `json:"node,omitempty"`
}

type FileFolderDeleteFieldInput struct {
	Where FolderWhere `json:"where"`
}

type FileRepositoryConnection struct {
	Edges      []*FileRepositoryEdge `json:"edges"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   PageInfo              `json:"pageInfo"`
}

type FileRepositoryEdge struct {
	Node   *Repository `json:"node"`
	Cursor string      `json:"cursor"`
}

type FileRepositoryFieldInput struct {
	Create  []*FileRepositoryCreateFieldInput  `json:"create,omitempty"`
	Connect []*FileRepositoryConnectFieldInput `json:"connect,omitempty"`
}

type FileRepositoryUpdateFieldInput struct {
	Create     []*FileRepositoryCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FileRepositoryConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FileRepositoryDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FileRepositoryUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FileRepositoryDeleteFieldInput      `json:"delete,omitempty"`
}

type FileRepositoryCreateFieldInput struct {
	Node RepositoryCreateInput `json:"node"`
}

type FileRepositoryConnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type FileRepositoryDisconnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type FileRepositoryUpdateConnectionInput struct {
	Where RepositoryWhere        `json:"where"`
	Node  *RepositoryUpdateInput `json:"node,omitempty"`
}

type FileRepositoryDeleteFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type Module struct {
	Id           string        `json:"id"`
	Name         string        `json:"name"`
	Path         string        `json:"path"`
	Language     *string       `json:"language,omitempty"`
	ImportPath   *string       `json:"importPath,omitempty"`
	Visibility   *string       `json:"visibility,omitempty"`
	Kind         *string       `json:"kind,omitempty"`
	StartingLine *int          `json:"startingLine,omitempty"`
	EndingLine   *int          `json:"endingLine,omitempty"`
	Functions    []*Function   `json:"functions,omitempty"`
	Classes      []*Class      `json:"classes,omitempty"`
	DependsOn    []*Module     `json:"dependsOn,omitempty"`
	DependedOnBy []*Module     `json:"dependedOnBy,omitempty"`
	ImportedBy   []*File       `json:"importedBy,omitempty"`
	Repository   []*Repository `json:"repository,omitempty"`
}

type ModuleCreateInput struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	Language     *string `json:"language,omitempty"`
	ImportPath   *string `json:"importPath,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	Kind         *string `json:"kind,omitempty"`
	StartingLine *int    `json:"startingLine,omitempty"`
	EndingLine   *int    `json:"endingLine,omitempty"`
}

type ModuleUpdateInput struct {
	Name         *string `json:"name,omitempty"`
	Path         *string `json:"path,omitempty"`
	Language     *string `json:"language,omitempty"`
	ImportPath   *string `json:"importPath,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	Kind         *string `json:"kind,omitempty"`
	StartingLine *int    `json:"startingLine,omitempty"`
	EndingLine   *int    `json:"endingLine,omitempty"`
}

type ModuleWhere struct {
	Id                   *string        `json:"id,omitempty"`
	IdNot                *string        `json:"id_NOT,omitempty"`
	IdIn                 []string       `json:"id_IN,omitempty"`
	IdNotIn              []string       `json:"id_NOT_IN,omitempty"`
	IdContains           *string        `json:"id_CONTAINS,omitempty"`
	IdStartsWith         *string        `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith           *string        `json:"id_ENDS_WITH,omitempty"`
	Name                 *string        `json:"name,omitempty"`
	NameNot              *string        `json:"name_NOT,omitempty"`
	NameIn               []string       `json:"name_IN,omitempty"`
	NameNotIn            []string       `json:"name_NOT_IN,omitempty"`
	NameGt               *string        `json:"name_GT,omitempty"`
	NameGte              *string        `json:"name_GTE,omitempty"`
	NameLt               *string        `json:"name_LT,omitempty"`
	NameLte              *string        `json:"name_LTE,omitempty"`
	NameContains         *string        `json:"name_CONTAINS,omitempty"`
	NameStartsWith       *string        `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith         *string        `json:"name_ENDS_WITH,omitempty"`
	Path                 *string        `json:"path,omitempty"`
	PathNot              *string        `json:"path_NOT,omitempty"`
	PathIn               []string       `json:"path_IN,omitempty"`
	PathNotIn            []string       `json:"path_NOT_IN,omitempty"`
	PathGt               *string        `json:"path_GT,omitempty"`
	PathGte              *string        `json:"path_GTE,omitempty"`
	PathLt               *string        `json:"path_LT,omitempty"`
	PathLte              *string        `json:"path_LTE,omitempty"`
	PathContains         *string        `json:"path_CONTAINS,omitempty"`
	PathStartsWith       *string        `json:"path_STARTS_WITH,omitempty"`
	PathEndsWith         *string        `json:"path_ENDS_WITH,omitempty"`
	Language             *string        `json:"language,omitempty"`
	LanguageNot          *string        `json:"language_NOT,omitempty"`
	LanguageIn           []string       `json:"language_IN,omitempty"`
	LanguageNotIn        []string       `json:"language_NOT_IN,omitempty"`
	LanguageGt           *string        `json:"language_GT,omitempty"`
	LanguageGte          *string        `json:"language_GTE,omitempty"`
	LanguageLt           *string        `json:"language_LT,omitempty"`
	LanguageLte          *string        `json:"language_LTE,omitempty"`
	LanguageContains     *string        `json:"language_CONTAINS,omitempty"`
	LanguageStartsWith   *string        `json:"language_STARTS_WITH,omitempty"`
	LanguageEndsWith     *string        `json:"language_ENDS_WITH,omitempty"`
	ImportPath           *string        `json:"importPath,omitempty"`
	ImportPathNot        *string        `json:"importPath_NOT,omitempty"`
	ImportPathIn         []string       `json:"importPath_IN,omitempty"`
	ImportPathNotIn      []string       `json:"importPath_NOT_IN,omitempty"`
	ImportPathGt         *string        `json:"importPath_GT,omitempty"`
	ImportPathGte        *string        `json:"importPath_GTE,omitempty"`
	ImportPathLt         *string        `json:"importPath_LT,omitempty"`
	ImportPathLte        *string        `json:"importPath_LTE,omitempty"`
	ImportPathContains   *string        `json:"importPath_CONTAINS,omitempty"`
	ImportPathStartsWith *string        `json:"importPath_STARTS_WITH,omitempty"`
	ImportPathEndsWith   *string        `json:"importPath_ENDS_WITH,omitempty"`
	Visibility           *string        `json:"visibility,omitempty"`
	VisibilityNot        *string        `json:"visibility_NOT,omitempty"`
	VisibilityIn         []string       `json:"visibility_IN,omitempty"`
	VisibilityNotIn      []string       `json:"visibility_NOT_IN,omitempty"`
	VisibilityGt         *string        `json:"visibility_GT,omitempty"`
	VisibilityGte        *string        `json:"visibility_GTE,omitempty"`
	VisibilityLt         *string        `json:"visibility_LT,omitempty"`
	VisibilityLte        *string        `json:"visibility_LTE,omitempty"`
	VisibilityContains   *string        `json:"visibility_CONTAINS,omitempty"`
	VisibilityStartsWith *string        `json:"visibility_STARTS_WITH,omitempty"`
	VisibilityEndsWith   *string        `json:"visibility_ENDS_WITH,omitempty"`
	Kind                 *string        `json:"kind,omitempty"`
	KindNot              *string        `json:"kind_NOT,omitempty"`
	KindIn               []string       `json:"kind_IN,omitempty"`
	KindNotIn            []string       `json:"kind_NOT_IN,omitempty"`
	KindGt               *string        `json:"kind_GT,omitempty"`
	KindGte              *string        `json:"kind_GTE,omitempty"`
	KindLt               *string        `json:"kind_LT,omitempty"`
	KindLte              *string        `json:"kind_LTE,omitempty"`
	KindContains         *string        `json:"kind_CONTAINS,omitempty"`
	KindStartsWith       *string        `json:"kind_STARTS_WITH,omitempty"`
	KindEndsWith         *string        `json:"kind_ENDS_WITH,omitempty"`
	StartingLine         *int           `json:"startingLine,omitempty"`
	StartingLineNot      *int           `json:"startingLine_NOT,omitempty"`
	StartingLineIn       []int          `json:"startingLine_IN,omitempty"`
	StartingLineNotIn    []int          `json:"startingLine_NOT_IN,omitempty"`
	StartingLineGt       *int           `json:"startingLine_GT,omitempty"`
	StartingLineGte      *int           `json:"startingLine_GTE,omitempty"`
	StartingLineLt       *int           `json:"startingLine_LT,omitempty"`
	StartingLineLte      *int           `json:"startingLine_LTE,omitempty"`
	EndingLine           *int           `json:"endingLine,omitempty"`
	EndingLineNot        *int           `json:"endingLine_NOT,omitempty"`
	EndingLineIn         []int          `json:"endingLine_IN,omitempty"`
	EndingLineNotIn      []int          `json:"endingLine_NOT_IN,omitempty"`
	EndingLineGt         *int           `json:"endingLine_GT,omitempty"`
	EndingLineGte        *int           `json:"endingLine_GTE,omitempty"`
	EndingLineLt         *int           `json:"endingLine_LT,omitempty"`
	EndingLineLte        *int           `json:"endingLine_LTE,omitempty"`
	AND                  []*ModuleWhere `json:"AND,omitempty"`
	OR                   []*ModuleWhere `json:"OR,omitempty"`
	NOT                  *ModuleWhere   `json:"NOT,omitempty"`
}

type ModuleSort struct {
	Id           *SortDirection `json:"id,omitempty"`
	Name         *SortDirection `json:"name,omitempty"`
	Path         *SortDirection `json:"path,omitempty"`
	Language     *SortDirection `json:"language,omitempty"`
	ImportPath   *SortDirection `json:"importPath,omitempty"`
	Visibility   *SortDirection `json:"visibility,omitempty"`
	Kind         *SortDirection `json:"kind,omitempty"`
	StartingLine *SortDirection `json:"startingLine,omitempty"`
	EndingLine   *SortDirection `json:"endingLine,omitempty"`
}

type ModulesConnection struct {
	Edges      []*ModuleEdge `json:"edges"`
	TotalCount int           `json:"totalCount"`
	PageInfo   PageInfo      `json:"pageInfo"`
}

type ModuleEdge struct {
	Node   *Module `json:"node"`
	Cursor string  `json:"cursor"`
}

type CreateModulesMutationResponse struct {
	Modules []*Module `json:"modules"`
}

type UpdateModulesMutationResponse struct {
	Modules []*Module `json:"modules"`
}

type ModuleFunctionsConnection struct {
	Edges      []*ModuleFunctionsEdge `json:"edges"`
	TotalCount int                    `json:"totalCount"`
	PageInfo   PageInfo               `json:"pageInfo"`
}

type ModuleFunctionsEdge struct {
	Node   *Function `json:"node"`
	Cursor string    `json:"cursor"`
}

type ModuleFunctionsFieldInput struct {
	Create  []*ModuleFunctionsCreateFieldInput  `json:"create,omitempty"`
	Connect []*ModuleFunctionsConnectFieldInput `json:"connect,omitempty"`
}

type ModuleFunctionsUpdateFieldInput struct {
	Create     []*ModuleFunctionsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ModuleFunctionsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ModuleFunctionsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ModuleFunctionsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ModuleFunctionsDeleteFieldInput      `json:"delete,omitempty"`
}

type ModuleFunctionsCreateFieldInput struct {
	Node FunctionCreateInput `json:"node"`
}

type ModuleFunctionsConnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type ModuleFunctionsDisconnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type ModuleFunctionsUpdateConnectionInput struct {
	Where FunctionWhere        `json:"where"`
	Node  *FunctionUpdateInput `json:"node,omitempty"`
}

type ModuleFunctionsDeleteFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type ModuleClassesConnection struct {
	Edges      []*ModuleClassesEdge `json:"edges"`
	TotalCount int                  `json:"totalCount"`
	PageInfo   PageInfo             `json:"pageInfo"`
}

type ModuleClassesEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type ModuleClassesFieldInput struct {
	Create  []*ModuleClassesCreateFieldInput  `json:"create,omitempty"`
	Connect []*ModuleClassesConnectFieldInput `json:"connect,omitempty"`
}

type ModuleClassesUpdateFieldInput struct {
	Create     []*ModuleClassesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ModuleClassesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ModuleClassesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ModuleClassesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ModuleClassesDeleteFieldInput      `json:"delete,omitempty"`
}

type ModuleClassesCreateFieldInput struct {
	Node ClassCreateInput `json:"node"`
}

type ModuleClassesConnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ModuleClassesDisconnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ModuleClassesUpdateConnectionInput struct {
	Where ClassWhere        `json:"where"`
	Node  *ClassUpdateInput `json:"node,omitempty"`
}

type ModuleClassesDeleteFieldInput struct {
	Where ClassWhere `json:"where"`
}

type ModuleDependsOnConnection struct {
	Edges      []*ModuleDependsOnEdge `json:"edges"`
	TotalCount int                    `json:"totalCount"`
	PageInfo   PageInfo               `json:"pageInfo"`
}

type ModuleDependsOnEdge struct {
	Node   *Module `json:"node"`
	Cursor string  `json:"cursor"`
}

type ModuleDependsOnFieldInput struct {
	Create  []*ModuleDependsOnCreateFieldInput  `json:"create,omitempty"`
	Connect []*ModuleDependsOnConnectFieldInput `json:"connect,omitempty"`
}

type ModuleDependsOnUpdateFieldInput struct {
	Create     []*ModuleDependsOnCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ModuleDependsOnConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ModuleDependsOnDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ModuleDependsOnUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ModuleDependsOnDeleteFieldInput      `json:"delete,omitempty"`
}

type ModuleDependsOnCreateFieldInput struct {
	Node ModuleCreateInput `json:"node"`
}

type ModuleDependsOnConnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ModuleDependsOnDisconnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ModuleDependsOnUpdateConnectionInput struct {
	Where ModuleWhere        `json:"where"`
	Node  *ModuleUpdateInput `json:"node,omitempty"`
}

type ModuleDependsOnDeleteFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ModuleDependedOnByConnection struct {
	Edges      []*ModuleDependedOnByEdge `json:"edges"`
	TotalCount int                       `json:"totalCount"`
	PageInfo   PageInfo                  `json:"pageInfo"`
}

type ModuleDependedOnByEdge struct {
	Node   *Module `json:"node"`
	Cursor string  `json:"cursor"`
}

type ModuleDependedOnByFieldInput struct {
	Create  []*ModuleDependedOnByCreateFieldInput  `json:"create,omitempty"`
	Connect []*ModuleDependedOnByConnectFieldInput `json:"connect,omitempty"`
}

type ModuleDependedOnByUpdateFieldInput struct {
	Create     []*ModuleDependedOnByCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ModuleDependedOnByConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ModuleDependedOnByDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ModuleDependedOnByUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ModuleDependedOnByDeleteFieldInput      `json:"delete,omitempty"`
}

type ModuleDependedOnByCreateFieldInput struct {
	Node ModuleCreateInput `json:"node"`
}

type ModuleDependedOnByConnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ModuleDependedOnByDisconnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ModuleDependedOnByUpdateConnectionInput struct {
	Where ModuleWhere        `json:"where"`
	Node  *ModuleUpdateInput `json:"node,omitempty"`
}

type ModuleDependedOnByDeleteFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ModuleImportedByConnection struct {
	Edges      []*ModuleImportedByEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type ModuleImportedByEdge struct {
	Node   *File  `json:"node"`
	Cursor string `json:"cursor"`
}

type ModuleImportedByFieldInput struct {
	Create  []*ModuleImportedByCreateFieldInput  `json:"create,omitempty"`
	Connect []*ModuleImportedByConnectFieldInput `json:"connect,omitempty"`
}

type ModuleImportedByUpdateFieldInput struct {
	Create     []*ModuleImportedByCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ModuleImportedByConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ModuleImportedByDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ModuleImportedByUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ModuleImportedByDeleteFieldInput      `json:"delete,omitempty"`
}

type ModuleImportedByCreateFieldInput struct {
	Node FileCreateInput `json:"node"`
}

type ModuleImportedByConnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type ModuleImportedByDisconnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type ModuleImportedByUpdateConnectionInput struct {
	Where FileWhere        `json:"where"`
	Node  *FileUpdateInput `json:"node,omitempty"`
}

type ModuleImportedByDeleteFieldInput struct {
	Where FileWhere `json:"where"`
}

type ModuleRepositoryConnection struct {
	Edges      []*ModuleRepositoryEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type ModuleRepositoryEdge struct {
	Node   *Repository `json:"node"`
	Cursor string      `json:"cursor"`
}

type ModuleRepositoryFieldInput struct {
	Create  []*ModuleRepositoryCreateFieldInput  `json:"create,omitempty"`
	Connect []*ModuleRepositoryConnectFieldInput `json:"connect,omitempty"`
}

type ModuleRepositoryUpdateFieldInput struct {
	Create     []*ModuleRepositoryCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ModuleRepositoryConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ModuleRepositoryDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ModuleRepositoryUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ModuleRepositoryDeleteFieldInput      `json:"delete,omitempty"`
}

type ModuleRepositoryCreateFieldInput struct {
	Node RepositoryCreateInput `json:"node"`
}

type ModuleRepositoryConnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type ModuleRepositoryDisconnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type ModuleRepositoryUpdateConnectionInput struct {
	Where RepositoryWhere        `json:"where"`
	Node  *RepositoryUpdateInput `json:"node,omitempty"`
}

type ModuleRepositoryDeleteFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type Function struct {
	Id                   string               `json:"id"`
	Name                 string               `json:"name"`
	Path                 string               `json:"path"`
	Language             *string              `json:"language,omitempty"`
	Signature            *string              `json:"signature,omitempty"`
	Visibility           *string              `json:"visibility,omitempty"`
	Source               *string              `json:"source,omitempty"`
	StartingLine         *int                 `json:"startingLine,omitempty"`
	EndingLine           *int                 `json:"endingLine,omitempty"`
	CyclomaticComplexity *int                 `json:"cyclomaticComplexity,omitempty"`
	Decorators           []string             `json:"decorators,omitempty"`
	Calls                []*Function          `json:"calls,omitempty"`
	CalledBy             []*Function          `json:"calledBy,omitempty"`
	ExternalCalls        []*ExternalReference `json:"externalCalls,omitempty"`
	Overrides            []*Function          `json:"overrides,omitempty"`
	OverriddenBy         []*Function          `json:"overriddenBy,omitempty"`
	DefinedIn            []*File              `json:"definedIn,omitempty"`
	Class                []*Class             `json:"class,omitempty"`
	Module               []*Module            `json:"module,omitempty"`
	Repository           []*Repository        `json:"repository,omitempty"`
}

type FunctionCreateInput struct {
	Name                 string   `json:"name"`
	Path                 string   `json:"path"`
	Language             *string  `json:"language,omitempty"`
	Signature            *string  `json:"signature,omitempty"`
	Visibility           *string  `json:"visibility,omitempty"`
	Source               *string  `json:"source,omitempty"`
	StartingLine         *int     `json:"startingLine,omitempty"`
	EndingLine           *int     `json:"endingLine,omitempty"`
	CyclomaticComplexity *int     `json:"cyclomaticComplexity,omitempty"`
	Decorators           []string `json:"decorators,omitempty"`
}

type FunctionUpdateInput struct {
	Name                 *string  `json:"name,omitempty"`
	Path                 *string  `json:"path,omitempty"`
	Language             *string  `json:"language,omitempty"`
	Signature            *string  `json:"signature,omitempty"`
	Visibility           *string  `json:"visibility,omitempty"`
	Source               *string  `json:"source,omitempty"`
	StartingLine         *int     `json:"startingLine,omitempty"`
	EndingLine           *int     `json:"endingLine,omitempty"`
	CyclomaticComplexity *int     `json:"cyclomaticComplexity,omitempty"`
	Decorators           []string `json:"decorators,omitempty"`
}

type FunctionWhere struct {
	Id                        *string          `json:"id,omitempty"`
	IdNot                     *string          `json:"id_NOT,omitempty"`
	IdIn                      []string         `json:"id_IN,omitempty"`
	IdNotIn                   []string         `json:"id_NOT_IN,omitempty"`
	IdContains                *string          `json:"id_CONTAINS,omitempty"`
	IdStartsWith              *string          `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith                *string          `json:"id_ENDS_WITH,omitempty"`
	Name                      *string          `json:"name,omitempty"`
	NameNot                   *string          `json:"name_NOT,omitempty"`
	NameIn                    []string         `json:"name_IN,omitempty"`
	NameNotIn                 []string         `json:"name_NOT_IN,omitempty"`
	NameGt                    *string          `json:"name_GT,omitempty"`
	NameGte                   *string          `json:"name_GTE,omitempty"`
	NameLt                    *string          `json:"name_LT,omitempty"`
	NameLte                   *string          `json:"name_LTE,omitempty"`
	NameContains              *string          `json:"name_CONTAINS,omitempty"`
	NameStartsWith            *string          `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith              *string          `json:"name_ENDS_WITH,omitempty"`
	Path                      *string          `json:"path,omitempty"`
	PathNot                   *string          `json:"path_NOT,omitempty"`
	PathIn                    []string         `json:"path_IN,omitempty"`
	PathNotIn                 []string         `json:"path_NOT_IN,omitempty"`
	PathGt                    *string          `json:"path_GT,omitempty"`
	PathGte                   *string          `json:"path_GTE,omitempty"`
	PathLt                    *string          `json:"path_LT,omitempty"`
	PathLte                   *string          `json:"path_LTE,omitempty"`
	PathContains              *string          `json:"path_CONTAINS,omitempty"`
	PathStartsWith            *string          `json:"path_STARTS_WITH,omitempty"`
	PathEndsWith              *string          `json:"path_ENDS_WITH,omitempty"`
	Language                  *string          `json:"language,omitempty"`
	LanguageNot               *string          `json:"language_NOT,omitempty"`
	LanguageIn                []string         `json:"language_IN,omitempty"`
	LanguageNotIn             []string         `json:"language_NOT_IN,omitempty"`
	LanguageGt                *string          `json:"language_GT,omitempty"`
	LanguageGte               *string          `json:"language_GTE,omitempty"`
	LanguageLt                *string          `json:"language_LT,omitempty"`
	LanguageLte               *string          `json:"language_LTE,omitempty"`
	LanguageContains          *string          `json:"language_CONTAINS,omitempty"`
	LanguageStartsWith        *string          `json:"language_STARTS_WITH,omitempty"`
	LanguageEndsWith          *string          `json:"language_ENDS_WITH,omitempty"`
	Signature                 *string          `json:"signature,omitempty"`
	SignatureNot              *string          `json:"signature_NOT,omitempty"`
	SignatureIn               []string         `json:"signature_IN,omitempty"`
	SignatureNotIn            []string         `json:"signature_NOT_IN,omitempty"`
	SignatureGt               *string          `json:"signature_GT,omitempty"`
	SignatureGte              *string          `json:"signature_GTE,omitempty"`
	SignatureLt               *string          `json:"signature_LT,omitempty"`
	SignatureLte              *string          `json:"signature_LTE,omitempty"`
	SignatureContains         *string          `json:"signature_CONTAINS,omitempty"`
	SignatureStartsWith       *string          `json:"signature_STARTS_WITH,omitempty"`
	SignatureEndsWith         *string          `json:"signature_ENDS_WITH,omitempty"`
	Visibility                *string          `json:"visibility,omitempty"`
	VisibilityNot             *string          `json:"visibility_NOT,omitempty"`
	VisibilityIn              []string         `json:"visibility_IN,omitempty"`
	VisibilityNotIn           []string         `json:"visibility_NOT_IN,omitempty"`
	VisibilityGt              *string          `json:"visibility_GT,omitempty"`
	VisibilityGte             *string          `json:"visibility_GTE,omitempty"`
	VisibilityLt              *string          `json:"visibility_LT,omitempty"`
	VisibilityLte             *string          `json:"visibility_LTE,omitempty"`
	VisibilityContains        *string          `json:"visibility_CONTAINS,omitempty"`
	VisibilityStartsWith      *string          `json:"visibility_STARTS_WITH,omitempty"`
	VisibilityEndsWith        *string          `json:"visibility_ENDS_WITH,omitempty"`
	Source                    *string          `json:"source,omitempty"`
	SourceNot                 *string          `json:"source_NOT,omitempty"`
	SourceIn                  []string         `json:"source_IN,omitempty"`
	SourceNotIn               []string         `json:"source_NOT_IN,omitempty"`
	SourceGt                  *string          `json:"source_GT,omitempty"`
	SourceGte                 *string          `json:"source_GTE,omitempty"`
	SourceLt                  *string          `json:"source_LT,omitempty"`
	SourceLte                 *string          `json:"source_LTE,omitempty"`
	SourceContains            *string          `json:"source_CONTAINS,omitempty"`
	SourceStartsWith          *string          `json:"source_STARTS_WITH,omitempty"`
	SourceEndsWith            *string          `json:"source_ENDS_WITH,omitempty"`
	StartingLine              *int             `json:"startingLine,omitempty"`
	StartingLineNot           *int             `json:"startingLine_NOT,omitempty"`
	StartingLineIn            []int            `json:"startingLine_IN,omitempty"`
	StartingLineNotIn         []int            `json:"startingLine_NOT_IN,omitempty"`
	StartingLineGt            *int             `json:"startingLine_GT,omitempty"`
	StartingLineGte           *int             `json:"startingLine_GTE,omitempty"`
	StartingLineLt            *int             `json:"startingLine_LT,omitempty"`
	StartingLineLte           *int             `json:"startingLine_LTE,omitempty"`
	EndingLine                *int             `json:"endingLine,omitempty"`
	EndingLineNot             *int             `json:"endingLine_NOT,omitempty"`
	EndingLineIn              []int            `json:"endingLine_IN,omitempty"`
	EndingLineNotIn           []int            `json:"endingLine_NOT_IN,omitempty"`
	EndingLineGt              *int             `json:"endingLine_GT,omitempty"`
	EndingLineGte             *int             `json:"endingLine_GTE,omitempty"`
	EndingLineLt              *int             `json:"endingLine_LT,omitempty"`
	EndingLineLte             *int             `json:"endingLine_LTE,omitempty"`
	CyclomaticComplexity      *int             `json:"cyclomaticComplexity,omitempty"`
	CyclomaticComplexityNot   *int             `json:"cyclomaticComplexity_NOT,omitempty"`
	CyclomaticComplexityIn    []int            `json:"cyclomaticComplexity_IN,omitempty"`
	CyclomaticComplexityNotIn []int            `json:"cyclomaticComplexity_NOT_IN,omitempty"`
	CyclomaticComplexityGt    *int             `json:"cyclomaticComplexity_GT,omitempty"`
	CyclomaticComplexityGte   *int             `json:"cyclomaticComplexity_GTE,omitempty"`
	CyclomaticComplexityLt    *int             `json:"cyclomaticComplexity_LT,omitempty"`
	CyclomaticComplexityLte   *int             `json:"cyclomaticComplexity_LTE,omitempty"`
	Decorators                *[]string        `json:"decorators,omitempty"`
	DecoratorsNot             *[]string        `json:"decorators_NOT,omitempty"`
	DecoratorsIn              [][]string       `json:"decorators_IN,omitempty"`
	DecoratorsNotIn           [][]string       `json:"decorators_NOT_IN,omitempty"`
	DecoratorsGt              *[]string        `json:"decorators_GT,omitempty"`
	DecoratorsGte             *[]string        `json:"decorators_GTE,omitempty"`
	DecoratorsLt              *[]string        `json:"decorators_LT,omitempty"`
	DecoratorsLte             *[]string        `json:"decorators_LTE,omitempty"`
	AND                       []*FunctionWhere `json:"AND,omitempty"`
	OR                        []*FunctionWhere `json:"OR,omitempty"`
	NOT                       *FunctionWhere   `json:"NOT,omitempty"`
}

type FunctionSort struct {
	Id                   *SortDirection `json:"id,omitempty"`
	Name                 *SortDirection `json:"name,omitempty"`
	Path                 *SortDirection `json:"path,omitempty"`
	Language             *SortDirection `json:"language,omitempty"`
	Signature            *SortDirection `json:"signature,omitempty"`
	Visibility           *SortDirection `json:"visibility,omitempty"`
	Source               *SortDirection `json:"source,omitempty"`
	StartingLine         *SortDirection `json:"startingLine,omitempty"`
	EndingLine           *SortDirection `json:"endingLine,omitempty"`
	CyclomaticComplexity *SortDirection `json:"cyclomaticComplexity,omitempty"`
	Decorators           *SortDirection `json:"decorators,omitempty"`
}

type FunctionsConnection struct {
	Edges      []*FunctionEdge `json:"edges"`
	TotalCount int             `json:"totalCount"`
	PageInfo   PageInfo        `json:"pageInfo"`
}

type FunctionEdge struct {
	Node   *Function `json:"node"`
	Cursor string    `json:"cursor"`
}

type CreateFunctionsMutationResponse struct {
	Functions []*Function `json:"functions"`
}

type UpdateFunctionsMutationResponse struct {
	Functions []*Function `json:"functions"`
}

type FunctionCallsConnection struct {
	Edges      []*FunctionCallsEdge `json:"edges"`
	TotalCount int                  `json:"totalCount"`
	PageInfo   PageInfo             `json:"pageInfo"`
}

type FunctionCallsEdge struct {
	Node       *Function       `json:"node"`
	Cursor     string          `json:"cursor"`
	Properties *CallProperties `json:"properties,omitempty"`
}

type FunctionCallsFieldInput struct {
	Create  []*FunctionCallsCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionCallsConnectFieldInput `json:"connect,omitempty"`
}

type FunctionCallsUpdateFieldInput struct {
	Create     []*FunctionCallsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionCallsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionCallsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionCallsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionCallsDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionCallsCreateFieldInput struct {
	Node FunctionCreateInput        `json:"node"`
	Edge *CallPropertiesCreateInput `json:"edge,omitempty"`
}

type FunctionCallsConnectFieldInput struct {
	Where FunctionWhere              `json:"where"`
	Edge  *CallPropertiesCreateInput `json:"edge,omitempty"`
}

type FunctionCallsDisconnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionCallsUpdateConnectionInput struct {
	Where FunctionWhere              `json:"where"`
	Node  *FunctionUpdateInput       `json:"node,omitempty"`
	Edge  *CallPropertiesUpdateInput `json:"edge,omitempty"`
}

type FunctionCallsDeleteFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionCalledByConnection struct {
	Edges      []*FunctionCalledByEdge `json:"edges"`
	TotalCount int                     `json:"totalCount"`
	PageInfo   PageInfo                `json:"pageInfo"`
}

type FunctionCalledByEdge struct {
	Node       *Function       `json:"node"`
	Cursor     string          `json:"cursor"`
	Properties *CallProperties `json:"properties,omitempty"`
}

type FunctionCalledByFieldInput struct {
	Create  []*FunctionCalledByCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionCalledByConnectFieldInput `json:"connect,omitempty"`
}

type FunctionCalledByUpdateFieldInput struct {
	Create     []*FunctionCalledByCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionCalledByConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionCalledByDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionCalledByUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionCalledByDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionCalledByCreateFieldInput struct {
	Node FunctionCreateInput        `json:"node"`
	Edge *CallPropertiesCreateInput `json:"edge,omitempty"`
}

type FunctionCalledByConnectFieldInput struct {
	Where FunctionWhere              `json:"where"`
	Edge  *CallPropertiesCreateInput `json:"edge,omitempty"`
}

type FunctionCalledByDisconnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionCalledByUpdateConnectionInput struct {
	Where FunctionWhere              `json:"where"`
	Node  *FunctionUpdateInput       `json:"node,omitempty"`
	Edge  *CallPropertiesUpdateInput `json:"edge,omitempty"`
}

type FunctionCalledByDeleteFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionExternalCallsConnection struct {
	Edges      []*FunctionExternalCallsEdge `json:"edges"`
	TotalCount int                          `json:"totalCount"`
	PageInfo   PageInfo                     `json:"pageInfo"`
}

type FunctionExternalCallsEdge struct {
	Node   *ExternalReference `json:"node"`
	Cursor string             `json:"cursor"`
}

type FunctionExternalCallsFieldInput struct {
	Create  []*FunctionExternalCallsCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionExternalCallsConnectFieldInput `json:"connect,omitempty"`
}

type FunctionExternalCallsUpdateFieldInput struct {
	Create     []*FunctionExternalCallsCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionExternalCallsConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionExternalCallsDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionExternalCallsUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionExternalCallsDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionExternalCallsCreateFieldInput struct {
	Node ExternalReferenceCreateInput `json:"node"`
}

type FunctionExternalCallsConnectFieldInput struct {
	Where ExternalReferenceWhere `json:"where"`
}

type FunctionExternalCallsDisconnectFieldInput struct {
	Where ExternalReferenceWhere `json:"where"`
}

type FunctionExternalCallsUpdateConnectionInput struct {
	Where ExternalReferenceWhere        `json:"where"`
	Node  *ExternalReferenceUpdateInput `json:"node,omitempty"`
}

type FunctionExternalCallsDeleteFieldInput struct {
	Where ExternalReferenceWhere `json:"where"`
}

type FunctionOverridesConnection struct {
	Edges      []*FunctionOverridesEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type FunctionOverridesEdge struct {
	Node   *Function `json:"node"`
	Cursor string    `json:"cursor"`
}

type FunctionOverridesFieldInput struct {
	Create  []*FunctionOverridesCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionOverridesConnectFieldInput `json:"connect,omitempty"`
}

type FunctionOverridesUpdateFieldInput struct {
	Create     []*FunctionOverridesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionOverridesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionOverridesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionOverridesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionOverridesDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionOverridesCreateFieldInput struct {
	Node FunctionCreateInput `json:"node"`
}

type FunctionOverridesConnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionOverridesDisconnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionOverridesUpdateConnectionInput struct {
	Where FunctionWhere        `json:"where"`
	Node  *FunctionUpdateInput `json:"node,omitempty"`
}

type FunctionOverridesDeleteFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionOverriddenByConnection struct {
	Edges      []*FunctionOverriddenByEdge `json:"edges"`
	TotalCount int                         `json:"totalCount"`
	PageInfo   PageInfo                    `json:"pageInfo"`
}

type FunctionOverriddenByEdge struct {
	Node   *Function `json:"node"`
	Cursor string    `json:"cursor"`
}

type FunctionOverriddenByFieldInput struct {
	Create  []*FunctionOverriddenByCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionOverriddenByConnectFieldInput `json:"connect,omitempty"`
}

type FunctionOverriddenByUpdateFieldInput struct {
	Create     []*FunctionOverriddenByCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionOverriddenByConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionOverriddenByDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionOverriddenByUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionOverriddenByDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionOverriddenByCreateFieldInput struct {
	Node FunctionCreateInput `json:"node"`
}

type FunctionOverriddenByConnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionOverriddenByDisconnectFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionOverriddenByUpdateConnectionInput struct {
	Where FunctionWhere        `json:"where"`
	Node  *FunctionUpdateInput `json:"node,omitempty"`
}

type FunctionOverriddenByDeleteFieldInput struct {
	Where FunctionWhere `json:"where"`
}

type FunctionDefinedInConnection struct {
	Edges      []*FunctionDefinedInEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type FunctionDefinedInEdge struct {
	Node   *File  `json:"node"`
	Cursor string `json:"cursor"`
}

type FunctionDefinedInFieldInput struct {
	Create  []*FunctionDefinedInCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionDefinedInConnectFieldInput `json:"connect,omitempty"`
}

type FunctionDefinedInUpdateFieldInput struct {
	Create     []*FunctionDefinedInCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionDefinedInConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionDefinedInDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionDefinedInUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionDefinedInDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionDefinedInCreateFieldInput struct {
	Node FileCreateInput `json:"node"`
}

type FunctionDefinedInConnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type FunctionDefinedInDisconnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type FunctionDefinedInUpdateConnectionInput struct {
	Where FileWhere        `json:"where"`
	Node  *FileUpdateInput `json:"node,omitempty"`
}

type FunctionDefinedInDeleteFieldInput struct {
	Where FileWhere `json:"where"`
}

type FunctionClassConnection struct {
	Edges      []*FunctionClassEdge `json:"edges"`
	TotalCount int                  `json:"totalCount"`
	PageInfo   PageInfo             `json:"pageInfo"`
}

type FunctionClassEdge struct {
	Node   *Class `json:"node"`
	Cursor string `json:"cursor"`
}

type FunctionClassFieldInput struct {
	Create  []*FunctionClassCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionClassConnectFieldInput `json:"connect,omitempty"`
}

type FunctionClassUpdateFieldInput struct {
	Create     []*FunctionClassCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionClassConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionClassDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionClassUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionClassDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionClassCreateFieldInput struct {
	Node ClassCreateInput `json:"node"`
}

type FunctionClassConnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type FunctionClassDisconnectFieldInput struct {
	Where ClassWhere `json:"where"`
}

type FunctionClassUpdateConnectionInput struct {
	Where ClassWhere        `json:"where"`
	Node  *ClassUpdateInput `json:"node,omitempty"`
}

type FunctionClassDeleteFieldInput struct {
	Where ClassWhere `json:"where"`
}

type FunctionModuleConnection struct {
	Edges      []*FunctionModuleEdge `json:"edges"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   PageInfo              `json:"pageInfo"`
}

type FunctionModuleEdge struct {
	Node   *Module `json:"node"`
	Cursor string  `json:"cursor"`
}

type FunctionModuleFieldInput struct {
	Create  []*FunctionModuleCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionModuleConnectFieldInput `json:"connect,omitempty"`
}

type FunctionModuleUpdateFieldInput struct {
	Create     []*FunctionModuleCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionModuleConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionModuleDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionModuleUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionModuleDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionModuleCreateFieldInput struct {
	Node ModuleCreateInput `json:"node"`
}

type FunctionModuleConnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type FunctionModuleDisconnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type FunctionModuleUpdateConnectionInput struct {
	Where ModuleWhere        `json:"where"`
	Node  *ModuleUpdateInput `json:"node,omitempty"`
}

type FunctionModuleDeleteFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type FunctionRepositoryConnection struct {
	Edges      []*FunctionRepositoryEdge `json:"edges"`
	TotalCount int                       `json:"totalCount"`
	PageInfo   PageInfo                  `json:"pageInfo"`
}

type FunctionRepositoryEdge struct {
	Node   *Repository `json:"node"`
	Cursor string      `json:"cursor"`
}

type FunctionRepositoryFieldInput struct {
	Create  []*FunctionRepositoryCreateFieldInput  `json:"create,omitempty"`
	Connect []*FunctionRepositoryConnectFieldInput `json:"connect,omitempty"`
}

type FunctionRepositoryUpdateFieldInput struct {
	Create     []*FunctionRepositoryCreateFieldInput      `json:"create,omitempty"`
	Connect    []*FunctionRepositoryConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*FunctionRepositoryDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*FunctionRepositoryUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*FunctionRepositoryDeleteFieldInput      `json:"delete,omitempty"`
}

type FunctionRepositoryCreateFieldInput struct {
	Node RepositoryCreateInput `json:"node"`
}

type FunctionRepositoryConnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type FunctionRepositoryDisconnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type FunctionRepositoryUpdateConnectionInput struct {
	Where RepositoryWhere        `json:"where"`
	Node  *RepositoryUpdateInput `json:"node,omitempty"`
}

type FunctionRepositoryDeleteFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type Repository struct {
	Id              string    `json:"id"`
	Name            string    `json:"name"`
	Path            *string   `json:"path,omitempty"`
	RemoteUrl       *string   `json:"remoteUrl,omitempty"`
	PrimaryLanguage *string   `json:"primaryLanguage,omitempty"`
	LastIndexed     DateTime  `json:"lastIndexed"`
	Folders         []*Folder `json:"folders,omitempty"`
	Files           []*File   `json:"files,omitempty"`
	Modules         []*Module `json:"modules,omitempty"`
}

type RepositoryCreateInput struct {
	Name            string   `json:"name"`
	Path            *string  `json:"path,omitempty"`
	RemoteUrl       *string  `json:"remoteUrl,omitempty"`
	PrimaryLanguage *string  `json:"primaryLanguage,omitempty"`
	LastIndexed     DateTime `json:"lastIndexed"`
}

type RepositoryUpdateInput struct {
	Name            *string   `json:"name,omitempty"`
	Path            *string   `json:"path,omitempty"`
	RemoteUrl       *string   `json:"remoteUrl,omitempty"`
	PrimaryLanguage *string   `json:"primaryLanguage,omitempty"`
	LastIndexed     *DateTime `json:"lastIndexed,omitempty"`
}

type RepositoryWhere struct {
	Id                        *string            `json:"id,omitempty"`
	IdNot                     *string            `json:"id_NOT,omitempty"`
	IdIn                      []string           `json:"id_IN,omitempty"`
	IdNotIn                   []string           `json:"id_NOT_IN,omitempty"`
	IdContains                *string            `json:"id_CONTAINS,omitempty"`
	IdStartsWith              *string            `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith                *string            `json:"id_ENDS_WITH,omitempty"`
	Name                      *string            `json:"name,omitempty"`
	NameNot                   *string            `json:"name_NOT,omitempty"`
	NameIn                    []string           `json:"name_IN,omitempty"`
	NameNotIn                 []string           `json:"name_NOT_IN,omitempty"`
	NameGt                    *string            `json:"name_GT,omitempty"`
	NameGte                   *string            `json:"name_GTE,omitempty"`
	NameLt                    *string            `json:"name_LT,omitempty"`
	NameLte                   *string            `json:"name_LTE,omitempty"`
	NameContains              *string            `json:"name_CONTAINS,omitempty"`
	NameStartsWith            *string            `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith              *string            `json:"name_ENDS_WITH,omitempty"`
	Path                      *string            `json:"path,omitempty"`
	PathNot                   *string            `json:"path_NOT,omitempty"`
	PathIn                    []string           `json:"path_IN,omitempty"`
	PathNotIn                 []string           `json:"path_NOT_IN,omitempty"`
	PathGt                    *string            `json:"path_GT,omitempty"`
	PathGte                   *string            `json:"path_GTE,omitempty"`
	PathLt                    *string            `json:"path_LT,omitempty"`
	PathLte                   *string            `json:"path_LTE,omitempty"`
	PathContains              *string            `json:"path_CONTAINS,omitempty"`
	PathStartsWith            *string            `json:"path_STARTS_WITH,omitempty"`
	PathEndsWith              *string            `json:"path_ENDS_WITH,omitempty"`
	RemoteUrl                 *string            `json:"remoteUrl,omitempty"`
	RemoteUrlNot              *string            `json:"remoteUrl_NOT,omitempty"`
	RemoteUrlIn               []string           `json:"remoteUrl_IN,omitempty"`
	RemoteUrlNotIn            []string           `json:"remoteUrl_NOT_IN,omitempty"`
	RemoteUrlGt               *string            `json:"remoteUrl_GT,omitempty"`
	RemoteUrlGte              *string            `json:"remoteUrl_GTE,omitempty"`
	RemoteUrlLt               *string            `json:"remoteUrl_LT,omitempty"`
	RemoteUrlLte              *string            `json:"remoteUrl_LTE,omitempty"`
	RemoteUrlContains         *string            `json:"remoteUrl_CONTAINS,omitempty"`
	RemoteUrlStartsWith       *string            `json:"remoteUrl_STARTS_WITH,omitempty"`
	RemoteUrlEndsWith         *string            `json:"remoteUrl_ENDS_WITH,omitempty"`
	PrimaryLanguage           *string            `json:"primaryLanguage,omitempty"`
	PrimaryLanguageNot        *string            `json:"primaryLanguage_NOT,omitempty"`
	PrimaryLanguageIn         []string           `json:"primaryLanguage_IN,omitempty"`
	PrimaryLanguageNotIn      []string           `json:"primaryLanguage_NOT_IN,omitempty"`
	PrimaryLanguageGt         *string            `json:"primaryLanguage_GT,omitempty"`
	PrimaryLanguageGte        *string            `json:"primaryLanguage_GTE,omitempty"`
	PrimaryLanguageLt         *string            `json:"primaryLanguage_LT,omitempty"`
	PrimaryLanguageLte        *string            `json:"primaryLanguage_LTE,omitempty"`
	PrimaryLanguageContains   *string            `json:"primaryLanguage_CONTAINS,omitempty"`
	PrimaryLanguageStartsWith *string            `json:"primaryLanguage_STARTS_WITH,omitempty"`
	PrimaryLanguageEndsWith   *string            `json:"primaryLanguage_ENDS_WITH,omitempty"`
	LastIndexed               *DateTime          `json:"lastIndexed,omitempty"`
	LastIndexedNot            *DateTime          `json:"lastIndexed_NOT,omitempty"`
	LastIndexedIn             []DateTime         `json:"lastIndexed_IN,omitempty"`
	LastIndexedNotIn          []DateTime         `json:"lastIndexed_NOT_IN,omitempty"`
	LastIndexedGt             *DateTime          `json:"lastIndexed_GT,omitempty"`
	LastIndexedGte            *DateTime          `json:"lastIndexed_GTE,omitempty"`
	LastIndexedLt             *DateTime          `json:"lastIndexed_LT,omitempty"`
	LastIndexedLte            *DateTime          `json:"lastIndexed_LTE,omitempty"`
	AND                       []*RepositoryWhere `json:"AND,omitempty"`
	OR                        []*RepositoryWhere `json:"OR,omitempty"`
	NOT                       *RepositoryWhere   `json:"NOT,omitempty"`
}

type RepositorySort struct {
	Id              *SortDirection `json:"id,omitempty"`
	Name            *SortDirection `json:"name,omitempty"`
	Path            *SortDirection `json:"path,omitempty"`
	RemoteUrl       *SortDirection `json:"remoteUrl,omitempty"`
	PrimaryLanguage *SortDirection `json:"primaryLanguage,omitempty"`
	LastIndexed     *SortDirection `json:"lastIndexed,omitempty"`
}

type RepositorysConnection struct {
	Edges      []*RepositoryEdge `json:"edges"`
	TotalCount int               `json:"totalCount"`
	PageInfo   PageInfo          `json:"pageInfo"`
}

type RepositoryEdge struct {
	Node   *Repository `json:"node"`
	Cursor string      `json:"cursor"`
}

type CreateRepositorysMutationResponse struct {
	Repositorys []*Repository `json:"repositorys"`
}

type UpdateRepositorysMutationResponse struct {
	Repositorys []*Repository `json:"repositorys"`
}

type RepositoryFoldersConnection struct {
	Edges      []*RepositoryFoldersEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type RepositoryFoldersEdge struct {
	Node   *Folder `json:"node"`
	Cursor string  `json:"cursor"`
}

type RepositoryFoldersFieldInput struct {
	Create  []*RepositoryFoldersCreateFieldInput  `json:"create,omitempty"`
	Connect []*RepositoryFoldersConnectFieldInput `json:"connect,omitempty"`
}

type RepositoryFoldersUpdateFieldInput struct {
	Create     []*RepositoryFoldersCreateFieldInput      `json:"create,omitempty"`
	Connect    []*RepositoryFoldersConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*RepositoryFoldersDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*RepositoryFoldersUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*RepositoryFoldersDeleteFieldInput      `json:"delete,omitempty"`
}

type RepositoryFoldersCreateFieldInput struct {
	Node FolderCreateInput `json:"node"`
}

type RepositoryFoldersConnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type RepositoryFoldersDisconnectFieldInput struct {
	Where FolderWhere `json:"where"`
}

type RepositoryFoldersUpdateConnectionInput struct {
	Where FolderWhere        `json:"where"`
	Node  *FolderUpdateInput `json:"node,omitempty"`
}

type RepositoryFoldersDeleteFieldInput struct {
	Where FolderWhere `json:"where"`
}

type RepositoryFilesConnection struct {
	Edges      []*RepositoryFilesEdge `json:"edges"`
	TotalCount int                    `json:"totalCount"`
	PageInfo   PageInfo               `json:"pageInfo"`
}

type RepositoryFilesEdge struct {
	Node   *File  `json:"node"`
	Cursor string `json:"cursor"`
}

type RepositoryFilesFieldInput struct {
	Create  []*RepositoryFilesCreateFieldInput  `json:"create,omitempty"`
	Connect []*RepositoryFilesConnectFieldInput `json:"connect,omitempty"`
}

type RepositoryFilesUpdateFieldInput struct {
	Create     []*RepositoryFilesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*RepositoryFilesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*RepositoryFilesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*RepositoryFilesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*RepositoryFilesDeleteFieldInput      `json:"delete,omitempty"`
}

type RepositoryFilesCreateFieldInput struct {
	Node FileCreateInput `json:"node"`
}

type RepositoryFilesConnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type RepositoryFilesDisconnectFieldInput struct {
	Where FileWhere `json:"where"`
}

type RepositoryFilesUpdateConnectionInput struct {
	Where FileWhere        `json:"where"`
	Node  *FileUpdateInput `json:"node,omitempty"`
}

type RepositoryFilesDeleteFieldInput struct {
	Where FileWhere `json:"where"`
}

type RepositoryModulesConnection struct {
	Edges      []*RepositoryModulesEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type RepositoryModulesEdge struct {
	Node   *Module `json:"node"`
	Cursor string  `json:"cursor"`
}

type RepositoryModulesFieldInput struct {
	Create  []*RepositoryModulesCreateFieldInput  `json:"create,omitempty"`
	Connect []*RepositoryModulesConnectFieldInput `json:"connect,omitempty"`
}

type RepositoryModulesUpdateFieldInput struct {
	Create     []*RepositoryModulesCreateFieldInput      `json:"create,omitempty"`
	Connect    []*RepositoryModulesConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*RepositoryModulesDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*RepositoryModulesUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*RepositoryModulesDeleteFieldInput      `json:"delete,omitempty"`
}

type RepositoryModulesCreateFieldInput struct {
	Node ModuleCreateInput `json:"node"`
}

type RepositoryModulesConnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type RepositoryModulesDisconnectFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type RepositoryModulesUpdateConnectionInput struct {
	Where ModuleWhere        `json:"where"`
	Node  *ModuleUpdateInput `json:"node,omitempty"`
}

type RepositoryModulesDeleteFieldInput struct {
	Where ModuleWhere `json:"where"`
}

type ExternalReference struct {
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	ImportPath string        `json:"importPath"`
	Repository []*Repository `json:"repository,omitempty"`
}

type ExternalReferenceCreateInput struct {
	Name       string `json:"name"`
	ImportPath string `json:"importPath"`
}

type ExternalReferenceUpdateInput struct {
	Name       *string `json:"name,omitempty"`
	ImportPath *string `json:"importPath,omitempty"`
}

type ExternalReferenceWhere struct {
	Id                   *string                   `json:"id,omitempty"`
	IdNot                *string                   `json:"id_NOT,omitempty"`
	IdIn                 []string                  `json:"id_IN,omitempty"`
	IdNotIn              []string                  `json:"id_NOT_IN,omitempty"`
	IdContains           *string                   `json:"id_CONTAINS,omitempty"`
	IdStartsWith         *string                   `json:"id_STARTS_WITH,omitempty"`
	IdEndsWith           *string                   `json:"id_ENDS_WITH,omitempty"`
	Name                 *string                   `json:"name,omitempty"`
	NameNot              *string                   `json:"name_NOT,omitempty"`
	NameIn               []string                  `json:"name_IN,omitempty"`
	NameNotIn            []string                  `json:"name_NOT_IN,omitempty"`
	NameGt               *string                   `json:"name_GT,omitempty"`
	NameGte              *string                   `json:"name_GTE,omitempty"`
	NameLt               *string                   `json:"name_LT,omitempty"`
	NameLte              *string                   `json:"name_LTE,omitempty"`
	NameContains         *string                   `json:"name_CONTAINS,omitempty"`
	NameStartsWith       *string                   `json:"name_STARTS_WITH,omitempty"`
	NameEndsWith         *string                   `json:"name_ENDS_WITH,omitempty"`
	ImportPath           *string                   `json:"importPath,omitempty"`
	ImportPathNot        *string                   `json:"importPath_NOT,omitempty"`
	ImportPathIn         []string                  `json:"importPath_IN,omitempty"`
	ImportPathNotIn      []string                  `json:"importPath_NOT_IN,omitempty"`
	ImportPathGt         *string                   `json:"importPath_GT,omitempty"`
	ImportPathGte        *string                   `json:"importPath_GTE,omitempty"`
	ImportPathLt         *string                   `json:"importPath_LT,omitempty"`
	ImportPathLte        *string                   `json:"importPath_LTE,omitempty"`
	ImportPathContains   *string                   `json:"importPath_CONTAINS,omitempty"`
	ImportPathStartsWith *string                   `json:"importPath_STARTS_WITH,omitempty"`
	ImportPathEndsWith   *string                   `json:"importPath_ENDS_WITH,omitempty"`
	AND                  []*ExternalReferenceWhere `json:"AND,omitempty"`
	OR                   []*ExternalReferenceWhere `json:"OR,omitempty"`
	NOT                  *ExternalReferenceWhere   `json:"NOT,omitempty"`
}

type ExternalReferenceSort struct {
	Id         *SortDirection `json:"id,omitempty"`
	Name       *SortDirection `json:"name,omitempty"`
	ImportPath *SortDirection `json:"importPath,omitempty"`
}

type ExternalReferencesConnection struct {
	Edges      []*ExternalReferenceEdge `json:"edges"`
	TotalCount int                      `json:"totalCount"`
	PageInfo   PageInfo                 `json:"pageInfo"`
}

type ExternalReferenceEdge struct {
	Node   *ExternalReference `json:"node"`
	Cursor string             `json:"cursor"`
}

type CreateExternalReferencesMutationResponse struct {
	ExternalReferences []*ExternalReference `json:"externalReferences"`
}

type UpdateExternalReferencesMutationResponse struct {
	ExternalReferences []*ExternalReference `json:"externalReferences"`
}

type ExternalReferenceRepositoryConnection struct {
	Edges      []*ExternalReferenceRepositoryEdge `json:"edges"`
	TotalCount int                                `json:"totalCount"`
	PageInfo   PageInfo                           `json:"pageInfo"`
}

type ExternalReferenceRepositoryEdge struct {
	Node   *Repository `json:"node"`
	Cursor string      `json:"cursor"`
}

type ExternalReferenceRepositoryFieldInput struct {
	Create  []*ExternalReferenceRepositoryCreateFieldInput  `json:"create,omitempty"`
	Connect []*ExternalReferenceRepositoryConnectFieldInput `json:"connect,omitempty"`
}

type ExternalReferenceRepositoryUpdateFieldInput struct {
	Create     []*ExternalReferenceRepositoryCreateFieldInput      `json:"create,omitempty"`
	Connect    []*ExternalReferenceRepositoryConnectFieldInput     `json:"connect,omitempty"`
	Disconnect []*ExternalReferenceRepositoryDisconnectFieldInput  `json:"disconnect,omitempty"`
	Update     []*ExternalReferenceRepositoryUpdateConnectionInput `json:"update,omitempty"`
	Delete     []*ExternalReferenceRepositoryDeleteFieldInput      `json:"delete,omitempty"`
}

type ExternalReferenceRepositoryCreateFieldInput struct {
	Node RepositoryCreateInput `json:"node"`
}

type ExternalReferenceRepositoryConnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type ExternalReferenceRepositoryDisconnectFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type ExternalReferenceRepositoryUpdateConnectionInput struct {
	Where RepositoryWhere        `json:"where"`
	Node  *RepositoryUpdateInput `json:"node,omitempty"`
}

type ExternalReferenceRepositoryDeleteFieldInput struct {
	Where RepositoryWhere `json:"where"`
}

type ClassMatchInput struct {
	Name         *string  `json:"name,omitempty"`
	Path         *string  `json:"path,omitempty"`
	Language     *string  `json:"language,omitempty"`
	Kind         *string  `json:"kind,omitempty"`
	Visibility   *string  `json:"visibility,omitempty"`
	Source       *string  `json:"source,omitempty"`
	StartingLine *int     `json:"startingLine,omitempty"`
	EndingLine   *int     `json:"endingLine,omitempty"`
	Decorators   []string `json:"decorators,omitempty"`
}

type ClassMergeInput struct {
	Match    *ClassMatchInput  `json:"match"`
	OnCreate *ClassCreateInput `json:"onCreate,omitempty"`
	OnMatch  *ClassUpdateInput `json:"onMatch,omitempty"`
}

type MergeClasssMutationResponse struct {
	Classs []*Class `json:"classs"`
}

type FolderMatchInput struct {
	Path        *string   `json:"path,omitempty"`
	LastUpdated *DateTime `json:"lastUpdated,omitempty"`
}

type FolderMergeInput struct {
	Match    *FolderMatchInput  `json:"match"`
	OnCreate *FolderCreateInput `json:"onCreate,omitempty"`
	OnMatch  *FolderUpdateInput `json:"onMatch,omitempty"`
}

type MergeFoldersMutationResponse struct {
	Folders []*Folder `json:"folders"`
}

type FileMatchInput struct {
	Path        *string   `json:"path,omitempty"`
	Filename    *string   `json:"filename,omitempty"`
	Language    *string   `json:"language,omitempty"`
	LineCount   *int      `json:"lineCount,omitempty"`
	LastUpdated *DateTime `json:"lastUpdated,omitempty"`
}

type FileMergeInput struct {
	Match    *FileMatchInput  `json:"match"`
	OnCreate *FileCreateInput `json:"onCreate,omitempty"`
	OnMatch  *FileUpdateInput `json:"onMatch,omitempty"`
}

type MergeFilesMutationResponse struct {
	Files []*File `json:"files"`
}

type ModuleMatchInput struct {
	Name         *string `json:"name,omitempty"`
	Path         *string `json:"path,omitempty"`
	Language     *string `json:"language,omitempty"`
	ImportPath   *string `json:"importPath,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	Kind         *string `json:"kind,omitempty"`
	StartingLine *int    `json:"startingLine,omitempty"`
	EndingLine   *int    `json:"endingLine,omitempty"`
}

type ModuleMergeInput struct {
	Match    *ModuleMatchInput  `json:"match"`
	OnCreate *ModuleCreateInput `json:"onCreate,omitempty"`
	OnMatch  *ModuleUpdateInput `json:"onMatch,omitempty"`
}

type MergeModulesMutationResponse struct {
	Modules []*Module `json:"modules"`
}

type FunctionMatchInput struct {
	Name                 *string  `json:"name,omitempty"`
	Path                 *string  `json:"path,omitempty"`
	Language             *string  `json:"language,omitempty"`
	Signature            *string  `json:"signature,omitempty"`
	Visibility           *string  `json:"visibility,omitempty"`
	Source               *string  `json:"source,omitempty"`
	StartingLine         *int     `json:"startingLine,omitempty"`
	EndingLine           *int     `json:"endingLine,omitempty"`
	CyclomaticComplexity *int     `json:"cyclomaticComplexity,omitempty"`
	Decorators           []string `json:"decorators,omitempty"`
}

type FunctionMergeInput struct {
	Match    *FunctionMatchInput  `json:"match"`
	OnCreate *FunctionCreateInput `json:"onCreate,omitempty"`
	OnMatch  *FunctionUpdateInput `json:"onMatch,omitempty"`
}

type MergeFunctionsMutationResponse struct {
	Functions []*Function `json:"functions"`
}

type RepositoryMatchInput struct {
	Name            *string   `json:"name,omitempty"`
	Path            *string   `json:"path,omitempty"`
	RemoteUrl       *string   `json:"remoteUrl,omitempty"`
	PrimaryLanguage *string   `json:"primaryLanguage,omitempty"`
	LastIndexed     *DateTime `json:"lastIndexed,omitempty"`
}

type RepositoryMergeInput struct {
	Match    *RepositoryMatchInput  `json:"match"`
	OnCreate *RepositoryCreateInput `json:"onCreate,omitempty"`
	OnMatch  *RepositoryUpdateInput `json:"onMatch,omitempty"`
}

type MergeRepositorysMutationResponse struct {
	Repositorys []*Repository `json:"repositorys"`
}

type ExternalReferenceMatchInput struct {
	Name       *string `json:"name,omitempty"`
	ImportPath *string `json:"importPath,omitempty"`
}

type ExternalReferenceMergeInput struct {
	Match    *ExternalReferenceMatchInput  `json:"match"`
	OnCreate *ExternalReferenceCreateInput `json:"onCreate,omitempty"`
	OnMatch  *ExternalReferenceUpdateInput `json:"onMatch,omitempty"`
}

type MergeExternalReferencesMutationResponse struct {
	ExternalReferences []*ExternalReference `json:"externalReferences"`
}

type ConnectClassMethodsInput struct {
	From *ClassWhere    `json:"from"`
	To   *FunctionWhere `json:"to"`
}

type ConnectInfo struct {
	RelationshipsCreated int `json:"relationshipsCreated"`
}

type ConnectClassInheritsInput struct {
	From *ClassWhere `json:"from"`
	To   *ClassWhere `json:"to"`
}

type ConnectClassInheritedByInput struct {
	From *ClassWhere `json:"from"`
	To   *ClassWhere `json:"to"`
}

type ConnectClassImplementsInput struct {
	From *ClassWhere `json:"from"`
	To   *ClassWhere `json:"to"`
}

type ConnectClassImplementedByInput struct {
	From *ClassWhere `json:"from"`
	To   *ClassWhere `json:"to"`
}

type ConnectClassDefinedInInput struct {
	From *ClassWhere `json:"from"`
	To   *FileWhere  `json:"to"`
}

type ConnectClassModuleInput struct {
	From *ClassWhere  `json:"from"`
	To   *ModuleWhere `json:"to"`
}

type ConnectClassRepositoryInput struct {
	From *ClassWhere      `json:"from"`
	To   *RepositoryWhere `json:"to"`
}

type ConnectFolderSubfoldersInput struct {
	From *FolderWhere `json:"from"`
	To   *FolderWhere `json:"to"`
}

type ConnectFolderFilesInput struct {
	From *FolderWhere `json:"from"`
	To   *FileWhere   `json:"to"`
}

type ConnectFolderParentFolderInput struct {
	From *FolderWhere `json:"from"`
	To   *FolderWhere `json:"to"`
}

type ConnectFolderRepositoryInput struct {
	From *FolderWhere     `json:"from"`
	To   *RepositoryWhere `json:"to"`
}

type ConnectFileFunctionsInput struct {
	From *FileWhere     `json:"from"`
	To   *FunctionWhere `json:"to"`
}

type ConnectFileClassesInput struct {
	From *FileWhere  `json:"from"`
	To   *ClassWhere `json:"to"`
}

type ConnectFileImportsInput struct {
	From *FileWhere   `json:"from"`
	To   *ModuleWhere `json:"to"`
}

type ConnectFileExternalImportsInput struct {
	From *FileWhere              `json:"from"`
	To   *ExternalReferenceWhere `json:"to"`
}

type ConnectFileFolderInput struct {
	From *FileWhere   `json:"from"`
	To   *FolderWhere `json:"to"`
}

type ConnectFileRepositoryInput struct {
	From *FileWhere       `json:"from"`
	To   *RepositoryWhere `json:"to"`
}

type ConnectModuleFunctionsInput struct {
	From *ModuleWhere   `json:"from"`
	To   *FunctionWhere `json:"to"`
}

type ConnectModuleClassesInput struct {
	From *ModuleWhere `json:"from"`
	To   *ClassWhere  `json:"to"`
}

type ConnectModuleDependsOnInput struct {
	From *ModuleWhere `json:"from"`
	To   *ModuleWhere `json:"to"`
}

type ConnectModuleDependedOnByInput struct {
	From *ModuleWhere `json:"from"`
	To   *ModuleWhere `json:"to"`
}

type ConnectModuleImportedByInput struct {
	From *ModuleWhere `json:"from"`
	To   *FileWhere   `json:"to"`
}

type ConnectModuleRepositoryInput struct {
	From *ModuleWhere     `json:"from"`
	To   *RepositoryWhere `json:"to"`
}

type ConnectFunctionCallsInput struct {
	From *FunctionWhere             `json:"from"`
	To   *FunctionWhere             `json:"to"`
	Edge *CallPropertiesCreateInput `json:"edge,omitempty"`
}

type ConnectFunctionCalledByInput struct {
	From *FunctionWhere             `json:"from"`
	To   *FunctionWhere             `json:"to"`
	Edge *CallPropertiesCreateInput `json:"edge,omitempty"`
}

type ConnectFunctionExternalCallsInput struct {
	From *FunctionWhere          `json:"from"`
	To   *ExternalReferenceWhere `json:"to"`
}

type ConnectFunctionOverridesInput struct {
	From *FunctionWhere `json:"from"`
	To   *FunctionWhere `json:"to"`
}

type ConnectFunctionOverriddenByInput struct {
	From *FunctionWhere `json:"from"`
	To   *FunctionWhere `json:"to"`
}

type ConnectFunctionDefinedInInput struct {
	From *FunctionWhere `json:"from"`
	To   *FileWhere     `json:"to"`
}

type ConnectFunctionClassInput struct {
	From *FunctionWhere `json:"from"`
	To   *ClassWhere    `json:"to"`
}

type ConnectFunctionModuleInput struct {
	From *FunctionWhere `json:"from"`
	To   *ModuleWhere   `json:"to"`
}

type ConnectFunctionRepositoryInput struct {
	From *FunctionWhere   `json:"from"`
	To   *RepositoryWhere `json:"to"`
}

type ConnectRepositoryFoldersInput struct {
	From *RepositoryWhere `json:"from"`
	To   *FolderWhere     `json:"to"`
}

type ConnectRepositoryFilesInput struct {
	From *RepositoryWhere `json:"from"`
	To   *FileWhere       `json:"to"`
}

type ConnectRepositoryModulesInput struct {
	From *RepositoryWhere `json:"from"`
	To   *ModuleWhere     `json:"to"`
}

type ConnectExternalReferenceRepositoryInput struct {
	From *ExternalReferenceWhere `json:"from"`
	To   *RepositoryWhere        `json:"to"`
}
