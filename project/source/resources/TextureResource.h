#pragma once

#include "../rendering/Texture.h"
#include "Resource.h"

struct TextureResourceHeader {
	unsigned short width;      /*< Texture width  */
	unsigned short height;     /*< Texture height */
	unsigned short format;     /*< Color format */
	unsigned short maxlod : 4; /*< Max LOD (0-10) */
	unsigned short minlod : 4; /*< Min LOD (0-10) */
	unsigned long _reserved : 16;
};

class TextureResource : public Resource {
public:
	TextureResource(void* base, unsigned int size) : Resource(base, size) {}
	Texture* Load();
	void Initialize() override;

private:
	TextureResourceHeader* header;
	Texture* internal;
	bool loaded;
};
