package llm

const (
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGoogle    ProviderType = "google"
	ProviderVertex    ProviderType = "vertex"

	ProviderOpenAI     ProviderType = "openai"
	ProviderGroq       ProviderType = "groq"
	ProviderTogether   ProviderType = "together"
	ProviderHyperbolic ProviderType = "hyperbolic"
	ProviderDeepseek   ProviderType = "deepseek"

	ProviderVoyage     ProviderType = "voyage"
	ProviderMixedBread ProviderType = "mixedbread"
)

// configs are user declared, here's some useful defaults
const (
	// LLM models

	// aka claude 3.6
	ConfigClaude3Dot6Sonnet = "claude-3.6-sonnet"
	// the traditional 3.5
	ConfigClaude3Dot5Sonnet = "claude-3.5-sonnet"

<<<<<<< HEAD
	ConfigGPT4Mini = "gpt-4-mini"
	ConfigGPT4o    = "gpt-4o"
=======
	ConfigGPT4Mini        = "gpt-4-mini"
	ConfigGPT4o           = "gpt-4o"
	ConfigOpenAIO1        = "oai-o1"
	ConfigOpenAIO1Mini    = "oai-o1-mini"
	ConfigOpenAIO1Preview = "oai-o1-preview"
>>>>>>> 8d31a77 (chore: update models, update versions)

	ConfigGroqLlama70B = "groq-llama-70b"
	ConfigGroqLlama8B  = "groq-llama-8b"
	ConfigGroqGemma9B  = "groq-gemma2-9b"
	ConfigGroqMixtral  = "groq-mixtral"

	ConfigTogetherGemma27B         = "together-gemma-27b"
	ConfigTogetherDeepseekCoder33B = "together-deepseek-coder-33b"

<<<<<<< HEAD
	ConfigGeminiFlash    = "gemini-flash"
	ConfigGeminiFlash8B  = "gemini-flash-8b"
	ConfigGeminiFlashExp = "gemini-flash-exp"
	ConfigGeminiPro      = "gemini-pro"
	ConfigGeminiProExp   = "gemini-pro-exp"
=======
	ConfigGemini1Dot5Flash = "gemini-flash"
	ConfigGemini1Dot5Pro   = "gemini-pro"
	ConfigGemini2Flash     = "gemini-2-flash"
>>>>>>> 8d31a77 (chore: update models, update versions)

	ConfigHyperbolicLlama405B     = "hyperbolic-llama-405b"
	ConfigHyperbolicLlama405BBase = "hyperbolic-llama-405b-base"
	ConfigHyperbolicLlama70B      = "hyperbolic-llama-70b"
	ConfigHyperbolicLlama8B       = "hyperbolic-llama-8b"

	ConfigDeepseekChat  = "deepseek-chat"
	ConfigDeepseekCoder = "deepseek-coder"

	// Vertex
	ConfigClaude3Dot5SonnetVertex = "claude-3.5-sonnet-vertex"
	ConfigLlama405BVertex         = "llama-405b-vertex"

	// Embedding models
	ConfigOpenAITextEmbedding3Small = "openai-text-embedding-3-small"
	ConfigOpenAITextEmbedding3Large = "openai-text-embedding-3-large"
	ConfigOpenAITextEmbeddingAda002 = "openai-text-embedding-ada-002"

	ConfigGeminiTextEmbedding4 = "gemini-text-embedding-004"

	ConfigMxbaiEmbedLargeV1    = "mxbai-embed-large-v1"
	ConfigVoyageLarge2Instruct = "voyage-large-2-instruct"
)

var configs = map[string]ModelConfig{
	ConfigClaude3Dot5Sonnet: {
		ProviderType:                     ProviderAnthropic,
		ModelName:                        "claude-3-5-sonnet-20240620",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  30000,
		CentiCentsPerMillionOutputTokens: 150000,
	},
	ConfigClaude3Dot5SonnetVertex: {
		ProviderType:                     ProviderVertex,
		ModelName:                        "claude-3-5-sonnet@20240620",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  30000,
		CentiCentsPerMillionOutputTokens: 150000,
	},
	ConfigLlama405BVertex: {
		ProviderType: ProviderVertex,
		ModelName:    "llama3-405b-instruct-maas",
		ModelType:    ModelTypeLLM,
		//	The Llama 3.1 API service is at no cost during public preview, and will be priced as per dollar-per-1M-tokens at GA.
	},
	ConfigClaude3Dot6Sonnet: {
		ProviderType: ProviderAnthropic,
		ModelName:    "claude-3-5-sonnet-20241022",
	},
	ConfigGPT4Mini: {
		ProviderType:                     ProviderOpenAI,
		ModelName:                        "gpt-4o-mini",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  1500,
		CentiCentsPerMillionOutputTokens: 6000,
	},
	ConfigGPT4o: {
<<<<<<< HEAD
		ProviderType:                     ProviderOpenAI,
		ModelName:                        "gpt-4o-2024-08-06",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  25000,
		CentiCentsPerMillionOutputTokens: 100000,
	},
	ConfigGroqLlama70B: {
		ProviderType:                     ProviderGroq,
		ModelName:                        "llama3-70b-8192",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  5900,
		CentiCentsPerMillionOutputTokens: 7900,
=======
		ProviderType: ProviderOpenAI,
		// NB 2025-01-08: 2024-08-06 remains 'latest', but there is also 'gpt-4o-2024-11-20'
		ModelName: "gpt-4o-2024-08-06",
	},
	ConfigOpenAIO1: {
		ProviderType: ProviderOpenAI,
		ModelName:    "o1",
	},
	ConfigOpenAIO1Mini: {
		ProviderType: ProviderOpenAI,
		ModelName:    "o1-mini",
	},
	ConfigOpenAIO1Preview: {
		ProviderType: ProviderOpenAI,
		ModelName:    "o1-preview",
	},
	ConfigGroqLlama70B: {
		ProviderType: ProviderGroq,
		ModelName:    "llama-3.3-70b-versatile",
>>>>>>> 8d31a77 (chore: update models, update versions)
	},
	ConfigGroqMixtral: {
		ProviderType:                     ProviderGroq,
		ModelName:                        "mixtral-8x7b-32768",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  2400,
		CentiCentsPerMillionOutputTokens: 2400,
	},
	ConfigGroqGemma9B: {
		ProviderType:                     ProviderGroq,
		ModelName:                        "gemma2-9b-it",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  2000,
		CentiCentsPerMillionOutputTokens: 2000,
	},
	ConfigGroqLlama8B: {
<<<<<<< HEAD
		ProviderType:                     ProviderGroq,
		ModelName:                        "llama3-8b-8192",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionInputTokens:  500,
		CentiCentsPerMillionOutputTokens: 800,
=======
		ProviderType: ProviderGroq,
		ModelName:    "llama-3.1-8b-instant",
>>>>>>> 8d31a77 (chore: update models, update versions)
	},
	ConfigTogetherGemma27B: {
		ProviderType:                     ProviderTogether,
		ModelName:                        "google/gemma-2-27b-it",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionOutputTokens: 8000,
		CentiCentsPerMillionInputTokens:  8000,
	},
	ConfigTogetherDeepseekCoder33B: {
		ProviderType:                     ProviderTogether,
		ModelName:                        "deepseek-ai/deepseek-coder-33b-instruct",
		ModelType:                        ModelTypeLLM,
		CentiCentsPerMillionOutputTokens: 8000,
		CentiCentsPerMillionInputTokens:  8000,
	},
	ConfigGemini1Dot5Flash: {
		ProviderType: ProviderGoogle,
		ModelName:    "gemini-1.5-flash",
		ModelType:    ModelTypeLLM,
		// assumes < 128k
		CentiCentsPerMillionOutputTokens: 3000,
		CentiCentsPerMillionInputTokens:  750,
	},
	ConfigGeminiFlashExp: {
		ProviderType: ProviderGoogle,
		ModelName:    "gemini-1.5-flash-exp-0827",
		ModelType:    ModelTypeLLM,
		// assumes < 128k
		CentiCentsPerMillionOutputTokens: 3000,
		CentiCentsPerMillionInputTokens:  750,
	},
	ConfigGeminiFlash8B: {
		ProviderType: ProviderGoogle,
		ModelName:    "gemini-1.5-flash-8b-exp-0827",
		ModelType:    ModelTypeLLM,
		// "pricing is TBD"
		CentiCentsPerMillionOutputTokens: 3000,
		CentiCentsPerMillionInputTokens:  750,
	},
	ConfigGemini1Dot5Pro: {
		ProviderType: ProviderGoogle,
		ModelName:    "gemini-1.5-pro",
		ModelType:    ModelTypeLLM,
		// assumes < 128k
		CentiCentsPerMillionOutputTokens: 35000,
		CentiCentsPerMillionInputTokens:  105000,
	},
	ConfigGemini2Flash: {
		ProviderType: ProviderGoogle,
<<<<<<< HEAD
		ModelName:    "gemini-1.5-pro-exp-0827",
		ModelType:    ModelTypeLLM,
		// assumes < 128k
		CentiCentsPerMillionOutputTokens: 35000,
		CentiCentsPerMillionInputTokens:  105000,
=======
		ModelName:    "gemini-2-flash-exp",
>>>>>>> 8d31a77 (chore: update models, update versions)
	},
	ConfigHyperbolicLlama405B: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-405B-Instruct",
		ModelType:    ModelTypeLLM,

		CentiCentsPerMillionInputTokens:  40000,
		CentiCentsPerMillionOutputTokens: 40000,
	},
	ConfigHyperbolicLlama405BBase: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-405B",
		ModelType:    ModelTypeLLM,

		CentiCentsPerMillionInputTokens:  40000,
		CentiCentsPerMillionOutputTokens: 40000,
	},
	ConfigHyperbolicLlama70B: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-70B-Instruct",
		ModelType:    ModelTypeLLM,

		CentiCentsPerMillionInputTokens:  4000,
		CentiCentsPerMillionOutputTokens: 4000,
	},
	ConfigHyperbolicLlama8B: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-8B-Instruct",
		ModelType:    ModelTypeLLM,

		CentiCentsPerMillionInputTokens:  1000,
		CentiCentsPerMillionOutputTokens: 1000,
	},

	ConfigDeepseekChat: {
		ProviderType: ProviderDeepseek,
		ModelName:    "deepseek-chat",
		ModelType:    ModelTypeLLM,

		// assume cache miss
		CentiCentsPerMillionInputTokens:  1400,
		CentiCentsPerMillionOutputTokens: 2800,
	},
	ConfigDeepseekCoder: {
		ProviderType: ProviderDeepseek,
		ModelName:    "deepseek-coder",
		ModelType:    ModelTypeLLM,

		// assume cache miss
		CentiCentsPerMillionInputTokens:  1400,
		CentiCentsPerMillionOutputTokens: 2800,
	},

	ConfigOpenAITextEmbedding3Small: {
		ProviderType: ProviderOpenAI,
		ModelName:    "text-embedding-3-small",
		ModelType:    ModelTypeEmbedding,
	},
	ConfigOpenAITextEmbedding3Large: {
		ProviderType: ProviderOpenAI,
		ModelName:    "text-embedding-3-large",
		ModelType:    ModelTypeEmbedding,
	},
	ConfigOpenAITextEmbeddingAda002: {
		ProviderType: ProviderOpenAI,
		ModelName:    "text-embedding-ada-002",
		ModelType:    ModelTypeEmbedding,
	},

	ConfigGeminiTextEmbedding4: {
		ProviderType: ProviderGoogle,
		ModelName:    "text-embedding-004",
		ModelType:    ModelTypeEmbedding,
	},

	ConfigMxbaiEmbedLargeV1: {
		ProviderType: ProviderMixedBread,
		ModelName:    "mxbai-embed-large-v1",
		ModelType:    ModelTypeEmbedding,
	},

	ConfigVoyageLarge2Instruct: {
		ProviderType: ProviderVoyage,
		ModelName:    "voyage-large-2-instruct",
		ModelType:    ModelTypeEmbedding,
	},
}
