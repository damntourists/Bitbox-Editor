package canvas

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/theme"
	"math"
	"sort"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/go-gl/mathgl/mgl32"
	"go.uber.org/zap"
)

type RenderPrimitive struct {
	*component.Component[*RenderPrimitive]
	UsePerspective bool
	PerspectiveFOV float32
	OrthoSize      float32
	Colormap       implot.Colormap

	primitives        map[string]*PrimitiveTopology
	selectedPrimitive string
}

func NewRenderPrimitive(id imgui.ID, primitiveName string) *RenderPrimitive {
	cmp := &RenderPrimitive{
		primitives:        Primitives,
		selectedPrimitive: primitiveName,
		UsePerspective:    true,
		PerspectiveFOV:    45.0,
		OrthoSize:         1.5,
	}

	cmp.Component = component.NewComponent[*RenderPrimitive](id, nil)
	cmp.SetLayoutBuilder(cmp)
	return cmp
}

func (c *RenderPrimitive) handleUpdate(cmd component.UpdateCmd) {
	if c.Component.HandleGlobalUpdate(cmd) {
		return
	}

	log.Warn(
		"LabelComponent unhandled update",
		zap.String("id", c.IDStr()),
		zap.Any("cmd", cmd),
	)
}

func (c *RenderPrimitive) SendUpdate(cmd component.UpdateCmd) {
	c.Component.SendUpdate(cmd)
}

func (c *RenderPrimitive) Layout() {
	c.Component.ProcessUpdates()
	c.LayoutFullscreen()
}

func (c *RenderPrimitive) LayoutFullscreen() {
	viewport := imgui.MainViewport()
	drawList := imgui.BackgroundDrawList()
	windowPos := viewport.Pos()
	windowSize := viewport.Size()

	c.renderPrimitive(drawList, windowPos, windowSize)
}

func (c *RenderPrimitive) renderPrimitive(drawList *imgui.DrawList, windowPos, windowSize imgui.Vec2) {
	center := imgui.NewVec2(
		windowPos.X+windowSize.X*0.5,
		windowPos.Y+windowSize.Y*0.5,
	)

	// Define the 3D transformations
	time := float32(imgui.Time()) / 48
	model := mgl32.HomogRotate3D(time, mgl32.Vec3{0.5, 1.0, 0.0})
	view := mgl32.LookAtV(mgl32.Vec3{0, 0, 2}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	var projection mgl32.Mat4
	aspect := float32(1.0)
	if windowSize.Y > 0 {
		aspect = windowSize.X / windowSize.Y
	}

	if c.UsePerspective {
		projection = mgl32.Perspective(mgl32.DegToRad(c.PerspectiveFOV), aspect, 0.1, 100.0)
	} else {
		projection = mgl32.Ortho(-c.OrthoSize*aspect, c.OrthoSize*aspect, -c.OrthoSize, c.OrthoSize, 0.1, 100.0)
	}

	mv := view.Mul4(model)
	currentPrimitive := c.primitives[c.selectedPrimitive]

	// Project vertices
	projectedVertices := make([]imgui.Vec2, len(currentPrimitive.Vertices))
	vertexDepths := make([]float32, len(currentPrimitive.Vertices))
	minDepth := float32(math.MaxFloat32)
	maxDepth := float32(-math.MaxFloat32)

	for i, vertex := range currentPrimitive.Vertices {
		v := vertex
		if currentPrimitive.Name == "Icosahedron" {
			v = v.Mul(0.3)
		}
		if currentPrimitive.Name == "Torus" {
			v = v.Mul(0.5)
		}

		v4 := mgl32.Vec4{v.X(), v.Y(), v.Z(), 1.0}
		viewSpaceVertex := mv.Mul4x1(v4)
		depth := viewSpaceVertex.Z()
		vertexDepths[i] = depth
		if depth < minDepth {
			minDepth = depth
		}
		if depth > maxDepth {
			maxDepth = depth
		}

		clipSpace := projection.Mul4x1(viewSpaceVertex)
		ndc := clipSpace.Vec3().Mul(1.0 / clipSpace.W())

		halfWidth := windowSize.X * 0.5
		halfHeight := windowSize.Y * 0.5
		screenX := center.X + (ndc.X() * halfWidth)
		screenY := center.Y - (ndc.Y() * halfHeight)

		projectedVertices[i] = imgui.NewVec2(screenX, screenY)
	}

	// Determine face visibility
	isFaceVisible := make([]bool, len(currentPrimitive.Faces))
	if !currentPrimitive.IsCullingDisabled {
		for i, face := range currentPrimitive.Faces {
			p0 := projectedVertices[face[0]]
			p1 := projectedVertices[face[1]]
			p2 := projectedVertices[face[2]]
			crossZ := (p1.X-p0.X)*(p2.Y-p0.Y) - (p1.Y-p0.Y)*(p2.X-p0.X)
			isFaceVisible[i] = crossZ < 0
		}
	}

	type drawableEdge struct {
		p1, p2        imgui.Vec2
		avgDepth      float32
		isVisible     bool
		colorU32      uint32
		ghostColorU32 uint32
	}

	// Draw the edges
	depthRange := maxDepth - minDepth
	edgesToDraw := make([]drawableEdge, 0, len(currentPrimitive.Edges))
	if depthRange < 0.0001 {
		depthRange = 1.0
	}
	const dashLength = 12.0
	const gapLength = 12.0
	for i, edge := range currentPrimitive.Edges {
		p1 := projectedVertices[edge[0]]
		p2 := projectedVertices[edge[1]]

		d1 := vertexDepths[edge[0]]
		d2 := vertexDepths[edge[1]]
		avgDepth := (d1 + d2) * 0.5

		normalizedDepth := (avgDepth - minDepth) / depthRange
		if normalizedDepth < 0.0 {
			normalizedDepth = 0.0
		}
		if normalizedDepth > 1.0 {
			normalizedDepth = 1.0
		}

		// TODO: Implement ability to invert colormap direction in settings.
		//invertedDepth := 1.0 - normalizedDepth
		//lineColorVec4 := implot.SampleColormapV(invertedDepth, theme.GetCurrentColormap())
		lineColorVec4 := implot.SampleColormapV(normalizedDepth, theme.GetCurrentColormap())
		lineColorU32 := imgui.ColorConvertFloat4ToU32(lineColorVec4)

		isVisible := true
		if !currentPrimitive.IsCullingDisabled {
			face1Idx := currentPrimitive.EdgeFaceMap[i][0]
			face2Idx := currentPrimitive.EdgeFaceMap[i][1]
			isVisible = isFaceVisible[face1Idx] || isFaceVisible[face2Idx]
		}

		ghostColorVec4 := lineColorVec4
		ghostColorVec4.W = ghostColorVec4.W * 0.75
		ghostColorU32 := imgui.ColorConvertFloat4ToU32(ghostColorVec4)

		edgesToDraw = append(edgesToDraw, drawableEdge{
			p1:            p1,
			p2:            p2,
			avgDepth:      avgDepth,
			isVisible:     isVisible,
			colorU32:      lineColorU32,
			ghostColorU32: ghostColorU32,
		})
	}

	// Sort the slice by depth (farthest to nearest)
	sort.Slice(edgesToDraw, func(i, j int) bool {
		return edgesToDraw[i].avgDepth < edgesToDraw[j].avgDepth
	})

	// Draw sorted edges
	for _, edge := range edgesToDraw {
		if edge.isVisible {
			drawList.AddLineV(edge.p1, edge.p2, edge.colorU32, 2.0)
		} else {
			// Lines for hidden edges are drawn as dashed lines
			addDashedLine(drawList, edge.p1, edge.p2, edge.ghostColorU32, 2.0, dashLength, gapLength)
		}
	}
}

func (c *RenderPrimitive) Destroy() {
	c.Component.Destroy()
}

func addDashedLine(drawList *imgui.DrawList, p1, p2 imgui.Vec2, col uint32, thickness float32, dashLen, gapLen float32) {
	vec := p2.Sub(p1)
	totalLen := float32(math.Sqrt(float64(vec.X*vec.X + vec.Y*vec.Y)))
	if totalLen < 0.0001 {
		return
	}
	dir := vec.Mul(1.0 / totalLen)
	currentPos := p1
	remainingLen := totalLen

	for remainingLen > 0.0001 {
		currentDashLen := float32(math.Min(float64(dashLen), float64(remainingLen)))
		dashEnd := currentPos.Add(dir.Mul(currentDashLen))
		drawList.AddLineV(currentPos, dashEnd, col, thickness)
		remainingLen -= currentDashLen
		if remainingLen <= 0.0001 {
			break
		}
		currentGapLen := float32(math.Min(float64(gapLen), float64(remainingLen)))
		currentPos = dashEnd.Add(dir.Mul(currentGapLen))
		remainingLen -= currentGapLen
	}
}
