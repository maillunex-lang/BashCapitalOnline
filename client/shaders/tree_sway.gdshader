shader_type canvas_item;

uniform float strength : hint_range(0.0, 0.08) = 0.025;
uniform float speed : hint_range(0.1, 4.0) = 1.4;

void vertex() {
	float wind = sin(TIME * speed + VERTEX.y * 0.15 + VERTEX.x * 0.05) * strength;
	VERTEX.x += wind * (1.0 - UV.y);
}
