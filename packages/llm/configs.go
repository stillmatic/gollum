package llm

const (
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGoogle    ProviderType = "google"

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
	ConfigClaude3Dot5Sonnet = "claude-3.5-sonnet"
	ConfigGPT4Mini          = "gpt-4-mini"
	ConfigGPT4o             = "gpt-4o"

	ConfigGroqLlama70B = "groq-llama-70b"
	ConfigGroqLlama8B  = "groq-llama-8b"
	ConfigGroqGemma9B  = "groq-gemma2-9b"
	ConfigGroqMixtral  = "groq-mixtral"

	ConfigTogetherGemma27B         = "together-gemma-27b"
	ConfigTogetherDeepseekCoder33B = "together-deepseek-coder-33b"

	ConfigGeminiFlash  = "gemini-flash"
	ConfigGeminiPro    = "gemini-pro"
	ConfigGeminiProExp = "gemini-pro-exp"

	ConfigHyperbolicLlama405B     = "hyperbolic-llama-405b"
	ConfigHyperbolicLlama405BBase = "hyperbolic-llama-405b-base"
	ConfigHyperbolicLlama70B      = "hyperbolic-llama-70b"
	ConfigHyperbolicLlama8B       = "hyperbolic-llama-8b"

	ConfigDeepseekChat  = "deepseek-chat"
	ConfigDeepseekCoder = "deepseek-coder"

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
		ProviderType: ProviderAnthropic,
		ModelName:    "claude-3-5-sonnet-20240620",
	},
	ConfigGPT4Mini: {
		ProviderType: ProviderOpenAI,
		ModelName:    "gpt-4o-mini",
	},
	ConfigGPT4o: {
		ProviderType: ProviderOpenAI,
		ModelName:    "gpt-4o-2024-08-06",
	},
	ConfigGroqLlama70B: {
		ProviderType: ProviderGroq,
		ModelName:    "llama3-70b-8192",
	},
	ConfigGroqMixtral: {
		ProviderType: ProviderGroq,
		ModelName:    "mixtral-8x7b-32768",
	},
	ConfigGroqGemma9B: {
		ProviderType: ProviderGroq,
		ModelName:    "gemma2-9b-it",
	},
	ConfigGroqLlama8B: {
		ProviderType: ProviderGroq,
		ModelName:    "llama3-8b-8192",
	},
	ConfigTogetherGemma27B: {
		ProviderType: ProviderTogether,
		ModelName:    "google/gemma-2-27b-it",
	},
	ConfigTogetherDeepseekCoder33B: {
		ProviderType: ProviderTogether,
		ModelName:    "deepseek-ai/deepseek-coder-33b-instruct",
	},
	ConfigGeminiFlash: {
		ProviderType: ProviderGoogle,
		ModelName:    "gemini-1.5-flash",
	},
	ConfigGeminiPro: {
		ProviderType: ProviderGoogle,
		ModelName:    "gemini-1.5-pro",
	},
	ConfigGeminiProExp: {
		ProviderType: ProviderGoogle,
		ModelName:    "gemini-1.5-pro-exp-0801",
	},
	ConfigHyperbolicLlama405B: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-405B-Instruct",
	},
	ConfigHyperbolicLlama405BBase: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-405B",
	},
	ConfigHyperbolicLlama70B: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-70B-Instruct",
	},
	ConfigHyperbolicLlama8B: {
		ProviderType: ProviderHyperbolic,
		ModelName:    "meta-llama/Meta-Llama-3.1-8B-Instruct",
	},

	ConfigDeepseekChat: {
		ProviderType: ProviderDeepseek,
		ModelName:    "deepseek-chat",
	},
	ConfigDeepseekCoder: {
		ProviderType: ProviderDeepseek,
		ModelName:    "deepseek-coder",
	},

	ConfigOpenAITextEmbedding3Small: {
		ProviderType: ProviderOpenAI,
		ModelName:    "text-embedding-3-small",
	},
	ConfigOpenAITextEmbedding3Large: {
		ProviderType: ProviderOpenAI,
		ModelName:    "text-embedding-3-large",
	},
	ConfigOpenAITextEmbeddingAda002: {
		ProviderType: ProviderOpenAI,
		ModelName:    "text-embedding-ada-002",
	},

	ConfigGeminiTextEmbedding4: {
		ProviderType: ProviderGoogle,
		ModelName:    "text-embedding-004",
	},

	ConfigMxbaiEmbedLargeV1: {
		ProviderType: ProviderMixedBread,
		ModelName:    "mxbai-embed-large-v1",
	},

	ConfigVoyageLarge2Instruct: {
		ProviderType: ProviderVoyage,
		ModelName:    "voyage-large-2-instruct",
	},
}
