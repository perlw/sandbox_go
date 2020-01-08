#version 330 core
in vec2 TexCoords;
out vec4 color;

uniform sampler2D text;
uniform vec3 textColor;
uniform int minBound;
uniform int maxBound;

void main() {
    float sampled = texture(text, TexCoords).r;
    float c = sampled * 255;
		c = (c - minBound) / (maxBound-minBound);

    if (c >= 0.5) {
        color = vec4(textColor, 1.0);
    } else {
        discard;
    }
}
