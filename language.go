package email_reply_parser //nolint:stylecheck,golint

var greetings = []string{
	// English
	`regards`,
	`best regards`,
	// French
	`meilleures salutations`,
	`cordialement`,
	// Polish
	`pozdrowienia`,
	`z poważaniem`,
	// Dutch
	`groeten`,
	`vriendelijke groeten`,
	// German
	`freundliche Grüße`,
	`grüße`,
	// Portuguese
	`cumprimentos`,
	`saudações`,
	// Norwegian
	`med vennlig hilsen`,
	`hilsen`,
	// Swedish
	`hälsningar`,
	`vänliga hälsningar`,
	// Danish
	`Med venlig hilsen`,
	`hilsen`,
	// Vietnamese
	`trân trọng`,
}

//nolint:gochecknoglobals
var labels = []string{
	// TODO: more languages
	"bel",
	"call",
	"tel",
	"email",
	"mail",
	"kvk",
	"vat",
	"btw",
}

//nolint:gochecknoglobals
var on = []string{
	// English
	`on`,
	// French
	`le`,
	// Polish
	`w dni`,
	// Dutch
	`op`,
	// German
	`am`,
	// Portuguese
	`em`,
	// Norwegian
	`på`,
	// Swedish, Danish
	`den`,
	// Vietnamese
	`vào`,
}

//nolint:gochecknoglobals
var wrote = []string{
	// English
	`wrote`, `sent`,
	// French
	`a écrit`,
	// Polish
	`napisał`,
	// Dutch
	`schreef`, `verzond`, `geschreven`,
	// German
	`schrieb`,
	// Portuguese
	`escreve`,
	// Norwegian, Swedish
	`skrev`,
	// Vietnamese
	`đã viết`,
}

var mailPrograms = []string{
	"iPhone",
	"Galaxy",
	"Samsung",
	"Mail",
	"Blackberry",
	"iPad",
	"Apple Mail",
	"Yahoo! Mail",
	"Outlook",
	"Outlook.com",
}

//nolint:gochecknoglobals
var sent = []string{
	// English
	`sent`,
	// French
	`envoyé`,
	// Polish
	`wysłane`,
	// Dutch
	`verzonden`,
	`verstuurd`,
	// German
	`geschickt`,
	// Portuguese
	`enviei`,
	// Norwegian, Swedish
	`sendt`, `skickas`,
	// Vietnamese
	`gởi`,
}

//nolint:gochecknoglobals
var forwarded = []string{
	// English
	`Forwarded`,
	// French
	`Transféré`,
	// Polish
	`Przekazane`,
	// Dutch
	`Doorgestuurd`,
	// German
	`Weitergeleitet`,
	// Portuguese
	`Encaminhado`, `Encaminhada`,
	// Norwegian, Swedish
	`Videresendt`, `Vidarebefordrad`,
	// Vietnamese
	`Chuyển tiếp`,
}
