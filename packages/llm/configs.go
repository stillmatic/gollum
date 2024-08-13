package llm

// configs that are user declared, here's some useful defaults
const (
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
		ModelName:    "gpt-4o",
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
}
