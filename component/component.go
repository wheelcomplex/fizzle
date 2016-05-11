// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package component

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/gombz"
	"github.com/tbogdala/groggy"
)

// ComponentMesh defines a mesh reference for a component and everything
// needed to draw it.
type ComponentMesh struct {
	// BinFile is a filepath should be relative to component file
	BinFile string

	// Textures specifies the texture files to load for mesh, relative
	// to the component file
	Textures []string

	// Offset is the location offset of the mesh in the component.
	Offset mgl.Vec3

	// Parent is the owning Component object
	Parent *Component

	// SrcMesh is the cached mesh data either from SrcFile or BinFile
	SrcMesh *gombz.Mesh
}

// ComponentChildRef defines a reference to another component JSON file
// so that Components can be built from other Component parts
type ComponentChildRef struct {
	File     string
	Location mgl.Vec3
}

// ComponentMaterial defines the material appearance of the component.
type ComponentMaterial struct {
	// ShaderName is the name of the shader program to use for rendering
	ShaderName string

	// Diffuse color for the material
	Diffuse mgl.Vec4
}

// CollisionRef specifies a collision object within the component
// (e.g. a collision cube for a wall).
// Note: right now it only supports AABB collisions.
type CollisionRef struct {
	Min  mgl.Vec3
	Max  mgl.Vec3
	Tags []string
}

// Component is the main structure for component JSON files.
type Component struct {
	// The name of the component
	Name string

	// Location is the location of the component.
	Location mgl.Vec3

	// All of the meshes that are part of this component
	Meshes []*ComponentMesh

	// The material description of the component
	Material *ComponentMaterial

	// ChildReferences can be specified to include other components
	// to be contained in this component.
	ChildReferences []*ComponentChildRef

	// Collision objects for the component
	Collisions []*CollisionRef

	// Properties is a map for client software's custom properties for the component.
	Properties map[string]string

	// this is the directory path for the component file if it was loaded
	// from JSON.
	componentDirPath string

	// this is the cached renerable object for the component that can
	// be used as a prototype.
	cachedRenderable *fizzle.Renderable
}

// Destroy will destroy the cached Renderable object if it exists.
func (c *Component) Destroy() {
	if c.cachedRenderable != nil {
		c.cachedRenderable.Destroy()
	}
}

// Clone makes a new component and then copies the members over
// to the new object. This means that Meshes, Collisions, ChildReferences, etc...
// are shared between the clones.
func (c *Component) Clone() *Component {
	clone := new(Component)

	// copy over all of the fields
	clone.Name = c.Name
	clone.Location = c.Location
	clone.Meshes = c.Meshes
	clone.ChildReferences = c.ChildReferences
	clone.Collisions = c.Collisions
	clone.Properties = c.Properties
	clone.Material = c.Material
	clone.componentDirPath = c.componentDirPath
	clone.cachedRenderable = c.cachedRenderable

	return clone
}

// GetRenderable will return the cached renderable object for the component
// or create one if it hasn't been made yet. The TextureManager is needed
// to resolve texture references.
func (c *Component) GetRenderable(tm *fizzle.TextureManager, shaders map[string]*fizzle.RenderShader) *fizzle.Renderable {
	// see if we have a cached renderable already created
	if c.cachedRenderable != nil {
		return c.cachedRenderable
	}

	// start by creating a renderable to hold all of the meshes
	group := fizzle.NewRenderable()
	group.IsGroup = true
	group.Location = c.Location

	// now create renderables for all of the meshes.
	// comnponents only create new render nodes for the meshs defined and
	// not for referenced components
	for _, compMesh := range c.Meshes {
		cmRenderable := createRenderableForMesh(tm, compMesh)
		group.AddChild(cmRenderable)

		// assign material properties if specified
		if c.Material != nil {
			cmRenderable.Core.DiffuseColor = c.Material.Diffuse
			cmRenderable.ShaderName = c.Material.ShaderName
			loadedShader, okay := shaders[c.Material.ShaderName]
			if okay {
				cmRenderable.Core.Shader = loadedShader
			}
		}

		// cache it for later
		c.cachedRenderable = cmRenderable
	}

	return group
}

// GetFullBinFilePath returns the full file path for the mesh binary file (gombz format).
func (cm *ComponentMesh) GetFullBinFilePath() string {
	return cm.Parent.componentDirPath + cm.BinFile
}

// GetFullTexturePath returns the full file path for the mesh texture.
func (cm *ComponentMesh) GetFullTexturePath(textureIndex int) string {
	return cm.Parent.componentDirPath + cm.Textures[textureIndex]
}

// GetVertices returns the vector slice containing the vertices for the mesh.
func (cm *ComponentMesh) GetVertices() ([]mgl.Vec3, error) {
	if cm.SrcMesh == nil {
		return nil, fmt.Errorf("No internal data present for component mesh to get vertices from.")
	}
	return cm.SrcMesh.Vertices, nil
}

// createRenderableForMesh does the work of creating the Renderable and putting all of
// the mesh data into VBOs.
func createRenderableForMesh(tm *fizzle.TextureManager, compMesh *ComponentMesh) *fizzle.Renderable {
	// create the new renderable
	r := fizzle.CreateFromGombz(compMesh.SrcMesh)

	// assign the texture
	if len(compMesh.Textures) > 0 {
		var okay bool
		r.Core.Tex0, okay = tm.GetTexture(compMesh.Textures[0])
		if !okay {
			groggy.Log("ERROR", "createRenderableForMesh failed to assign a texture gl id for %s.", compMesh.Textures[0])
		}
	}

	return r
}
