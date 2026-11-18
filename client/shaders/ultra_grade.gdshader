shader_type canvas_item;

uniform float vignette : hint_range(0.0, 1.0) = 0.35;
uniform float warmth : hint_range(-0.2, 0.2) = 0.04;
uniform float bloom_fake : hint_range(0.0, 0.4) = 0.12;
uniform float time_scale = 1.0;

void fragment() {
	vec4 col = texture(TEXTURE, UV);
	vec2 uv = UV - vec2(0.5);
	float vig = smoothstep(0.9, 0.2, length(uv) * (0.85 + vignette));
	col.rgb *= vig;
	col.r += warmth;
	col.b -= warmth * 0.5;
	// subtle animated light sweep for ultra
	float sweep = sin((UV.x + UV.y) * 6.0 + TIME * time_scale) * 0.015 * bloom_fake * 10.0;
	col.rgb += vec3(sweep);
	COLOR = col;
}
