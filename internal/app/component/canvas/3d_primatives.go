package canvas

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// PrimitiveTopology - Generic struct for any 3D primitive
type PrimitiveTopology struct {
	Name              string
	IsCullingDisabled bool
	Vertices          []mgl32.Vec3
	Edges             [][2]int
	Faces             [][4]int
	EdgeFaceMap       [][2]int
}

var (
	// Golden ratio for Icosahedron
	gr = (1.0 + float32(math.Sqrt(5.0))) / 2.0

	Primitives = make(map[string]*PrimitiveTopology)
)

func init() {
	// Torus
	torusMainSeg := 20
	torusTubeSeg := 12
	torusMainRad := 1.0
	torusTubeRad := 0.4
	vTorus, eTorus, fTorus, efTorus := GenerateTorusData(
		torusMainSeg,
		torusTubeSeg,
		float32(torusMainRad),
		float32(torusTubeRad),
	)
	Primitives["torus"] = &PrimitiveTopology{
		Name:              "Torus",
		IsCullingDisabled: false,
		Vertices:          vTorus,
		Edges:             eTorus,
		Faces:             fTorus,
		EdgeFaceMap:       efTorus,
	}

	// Sphere
	sphereStacks := 16
	sphereSectors := 24
	sphereRad := 1.0
	vSphere, eSphere, fSphere, efSphere := GenerateSphereData(
		sphereStacks,
		sphereSectors,
		float32(sphereRad),
	)
	Primitives["sphere"] = &PrimitiveTopology{
		Name:              "Sphere",
		IsCullingDisabled: false,
		Vertices:          vSphere,
		Edges:             eSphere,
		Faces:             fSphere,
		EdgeFaceMap:       efSphere,
	}

	// Cylinder
	cylSectors := 24
	cylHeight := 2.0
	cylRad := 1.0
	vCyl, eCyl, fCyl, efCyl := GenerateCylinderData(
		cylSectors,
		float32(cylHeight),
		float32(cylRad),
	)
	Primitives["cylinder"] = &PrimitiveTopology{
		Name:              "Cylinder",
		IsCullingDisabled: false,
		Vertices:          vCyl,
		Edges:             eCyl,
		Faces:             fCyl,
		EdgeFaceMap:       efCyl,
	}

	Primitives["cube"] = &PrimitiveTopology{
		Name: "Cube",
		Vertices: []mgl32.Vec3{
			{-0.5, -0.5, -0.5},
			{0.5, -0.5, -0.5},
			{0.5, 0.5, -0.5},
			{-0.5, 0.5, -0.5},
			{-0.5, -0.5, 0.5},
			{0.5, -0.5, 0.5},
			{0.5, 0.5, 0.5},
			{-0.5, 0.5, 0.5},
		},
		Edges: [][2]int{
			{0, 1}, {1, 2}, {2, 3}, {3, 0},
			{4, 5}, {5, 6}, {6, 7}, {7, 4},
			{0, 4}, {1, 5}, {2, 6}, {3, 7},
		},
		Faces: [][4]int{
			{4, 5, 6, 7},
			{3, 2, 1, 0},
			{3, 7, 6, 2},
			{0, 1, 5, 4},
			{1, 2, 6, 5},
			{0, 4, 7, 3},
		},
		EdgeFaceMap: [][2]int{
			{1, 3}, {1, 4}, {1, 2}, {1, 5}, // Edges 0-3
			{0, 3}, {0, 4}, {0, 2}, {0, 5}, // Edges 4-7
			{5, 3}, {4, 3}, {4, 2}, {5, 2}, // Edges 8-11
		},
	}

	Primitives["pyramid"] = &PrimitiveTopology{
		Name: "Pyramid",
		Vertices: []mgl32.Vec3{
			{-0.5, 0.0, -0.5}, // 0: Base
			{0.5, 0.0, -0.5},  // 1: Base
			{0.5, 0.0, 0.5},   // 2: Base
			{-0.5, 0.0, 0.5},  // 3: Base
			{0.0, 1.0, 0.0},   // 4: Apex
		},
		Edges: [][2]int{
			{0, 1}, {1, 2}, {2, 3}, {3, 0}, // Base
			{0, 4}, {1, 4}, {2, 4}, {3, 4}, // Sides
		},
		Faces: [][4]int{
			// Note: For triangles, we repeat the last vertex.
			{3, 2, 1, 0}, // 0: Bottom (Quad)
			{0, 1, 4, 4}, // 1: Back (Tri)
			{1, 2, 4, 4}, // 2: Right (Tri)
			{2, 3, 4, 4}, // 3: Front (Tri)
			{3, 0, 4, 4}, // 4: Left (Tri)
		},
		EdgeFaceMap: [][2]int{
			{0, 1}, {0, 2}, {0, 3}, {0, 4}, // Edges 0-3
			{1, 4}, {1, 2}, {2, 3}, {3, 4}, // Edges 4-7
		},
	}

	Primitives["tetrahedron"] = &PrimitiveTopology{
		Name: "Tetrahedron",
		Vertices: []mgl32.Vec3{
			{0.5, 0.5, 0.5},
			{-0.5, -0.5, 0.5},
			{-0.5, 0.5, -0.5},
			{0.5, -0.5, -0.5},
		},
		Edges: [][2]int{
			{0, 1}, {0, 2}, {0, 3}, // 0, 1, 2
			{1, 2}, {1, 3}, // 3, 4
			{2, 3}, // 5
		},
		Faces: [][4]int{
			{0, 2, 1, 1}, // 0
			{0, 1, 3, 3}, // 1
			{0, 3, 2, 2}, // 2
			{1, 2, 3, 3}, // 3
		},
		EdgeFaceMap: [][2]int{
			{0, 1}, // Edge 0
			{0, 2}, // Edge 1
			{1, 2}, // Edge 2
			{0, 3}, // Edge 3
			{1, 3}, // Edge 4
			{2, 3}, // Edge 5
		},
	}

	Primitives["octahedron"] = &PrimitiveTopology{
		Name: "Octahedron",
		Vertices: []mgl32.Vec3{
			{1, 0, 0},  // 0: X+
			{-1, 0, 0}, // 1: X-
			{0, 1, 0},  // 2: Y+ (Top)
			{0, -1, 0}, // 3: Y- (Bottom)
			{0, 0, 1},  // 4: Z+
			{0, 0, -1}, // 5: Z-
		},
		Edges: [][2]int{
			{0, 4}, {4, 1}, {1, 5}, {5, 0}, // 0-3: Equator (X-Z plane)
			{2, 0}, {2, 4}, {2, 1}, {2, 5}, // 4-7: Top cone
			{3, 0}, {3, 4}, {3, 1}, {3, 5}, // 8-11: Bottom cone
		},
		Faces: [][4]int{
			{2, 0, 4, 4}, // 0: Top-Front-Right
			{2, 4, 1, 1}, // 1: Top-Back-Right
			{2, 1, 5, 5}, // 2: Top-Back-Left
			{2, 5, 0, 0}, // 3: Top-Front-Left
			{3, 4, 0, 0}, // 4: Bottom-Front-Right
			{3, 1, 4, 4}, // 5: Bottom-Back-Right
			{3, 5, 1, 1}, // 6: Bottom-Back-Left
			{3, 0, 5, 5}, // 7: Bottom-Front-Left
		},
		EdgeFaceMap: [][2]int{
			{0, 4}, // Edge 0
			{1, 5}, // Edge 1
			{2, 6}, // Edge 2
			{3, 7}, // Edge 3
			{0, 3}, // Edge 4
			{0, 1}, // Edge 5
			{1, 2}, // Edge 6
			{2, 3}, // Edge 7
			{4, 7}, // Edge 8
			{4, 5}, // Edge 9
			{5, 6}, // Edge 10
			{6, 7}, // Edge 11
		},
	}

	Primitives["icosahedron"] = &PrimitiveTopology{
		Name:              "Icosahedron",
		IsCullingDisabled: false,
		Vertices: []mgl32.Vec3{
			{-1, gr, 0}, {1, gr, 0}, {-1, -gr, 0}, {1, -gr, 0},
			{0, -1, gr}, {0, 1, gr}, {0, -1, -gr}, {0, 1, -gr},
			{gr, 0, -1}, {gr, 0, 1}, {-gr, 0, -1}, {-gr, 0, 1},
		},
		Edges: [][2]int{
			{0, 11}, {0, 5}, {0, 7}, {0, 10}, {0, 1}, {1, 5}, {1, 7}, {1, 8}, {1, 9}, {2, 3}, {2, 4},
			{2, 6}, {2, 10}, {2, 11}, {3, 4}, {3, 6}, {3, 8}, {3, 9}, {4, 5}, {4, 11}, {5, 9}, {5, 11},
			{6, 7}, {6, 8}, {6, 10}, {7, 8}, {7, 10}, {8, 9}, {4, 9}, {10, 11},
		},
		Faces: [][4]int{
			{0, 11, 5, 5}, {0, 5, 1, 1}, {0, 1, 7, 7}, {0, 7, 10, 10}, {0, 10, 11, 11},
			{1, 5, 9, 9}, {5, 11, 4, 4}, {11, 10, 2, 2}, {10, 7, 6, 6}, {7, 1, 8, 8},
			{3, 9, 4, 4}, {3, 4, 2, 2}, {3, 2, 6, 6}, {3, 6, 8, 8}, {3, 8, 9, 9},
			{4, 9, 5, 5}, {2, 4, 11, 11}, {6, 2, 10, 10}, {8, 6, 7, 7}, {9, 8, 1, 1},
		},
		EdgeFaceMap: [][2]int{
			{0, 4},   // Edge 0: {0, 11}
			{0, 1},   // Edge 1: {0, 5}
			{2, 3},   // Edge 2: {0, 7}
			{3, 4},   // Edge 3: {0, 10}
			{1, 2},   // Edge 4: {0, 1}
			{1, 5},   // Edge 5: {1, 5}
			{2, 9},   // Edge 6: {1, 7}
			{9, 19},  // Edge 7: {1, 8}
			{5, 19},  // Edge 8: {1, 9}
			{11, 12}, // Edge 9: {2, 3}
			{11, 16}, // Edge 10: {2, 4}
			{12, 17}, // Edge 11: {2, 6}
			{7, 17},  // Edge 12: {2, 10}
			{7, 16},  // Edge 13: {2, 11}
			{10, 11}, // Edge 14: {3, 4}
			{12, 13}, // Edge 15: {3, 6}
			{13, 14}, // Edge 16: {3, 8}
			{10, 14}, // Edge 17: {3, 9}
			{6, 15},  // Edge 18: {4, 5}
			{6, 16},  // Edge 19: {4, 11}
			{5, 15},  // Edge 20: {5, 9}
			{0, 6},   // Edge 21: {5, 11}
			{8, 18},  // Edge 22: {6, 7}
			{13, 18}, // Edge 23: {6, 8}
			{8, 17},  // Edge 24: {6, 10}
			{9, 18},  // Edge 25: {7, 8}
			{3, 8},   // Edge 26: {7, 10}
			{14, 19}, // Edge 27: {8, 9}
			{10, 15}, // Edge 28: {4, 9} (This is the corrected edge)
			{4, 7},   // Edge 29: {10, 11}
		},
	}
}

func GenerateSphereData(stacks, sectors int, radius float32) (
	vertices []mgl32.Vec3,
	edges [][2]int,
	faces [][4]int,
	edgeFaceMap [][2]int,
) {
	// Add top pole
	vertices = append(vertices, mgl32.Vec3{0, radius, 0})
	topPoleIndex := 0

	// Add middle vertices
	for i := 1; i < stacks; i++ {
		// Latitude
		phi := math.Pi * (float64(i) / float64(stacks))
		for j := 0; j < sectors; j++ {
			// Longitude
			theta := 2.0 * math.Pi * (float64(j) / float64(sectors))
			x := radius * float32(math.Sin(phi)*math.Cos(theta))
			y := radius * float32(math.Cos(phi))
			z := radius * float32(math.Sin(phi)*math.Sin(theta))
			vertices = append(vertices, mgl32.Vec3{x, y, z})
		}
	}

	// Add bottom pole
	vertices = append(vertices, mgl32.Vec3{0, -radius, 0})
	bottomPoleIndex := len(vertices) - 1

	getIndex := func(i, j int) int {
		return 1 + i*sectors + (j % sectors)
	}

	// Generate faces
	edgeToFacesMap := make(map[[2]int][]int)
	normalizeAndAdd := func(v1, v2, faceIndex int) {
		key := [2]int{v1, v2}
		if v1 > v2 {
			key = [2]int{v2, v1}
		}
		edgeToFacesMap[key] = append(edgeToFacesMap[key], faceIndex)
	}

	faceIndex := 0
	// Generate top cap faces and edges
	for j := 0; j < sectors; j++ {
		p1 := getIndex(0, j)
		p2 := getIndex(0, j+1)
		faces = append(faces, [4]int{topPoleIndex, p2, p1, p1})
		normalizeAndAdd(topPoleIndex, p1, faceIndex)
		normalizeAndAdd(p1, p2, faceIndex)
		normalizeAndAdd(p2, topPoleIndex, faceIndex)
		faceIndex++
	}

	// Generate middle stack faces and edges
	for i := 0; i < stacks-2; i++ {
		for j := 0; j < sectors; j++ {
			p0 := getIndex(i, j)
			p1 := getIndex(i, j+1)
			p2 := getIndex(i+1, j+1)
			p3 := getIndex(i+1, j)
			faces = append(faces, [4]int{p0, p1, p2, p3})
			normalizeAndAdd(p0, p1, faceIndex)
			normalizeAndAdd(p1, p2, faceIndex)
			normalizeAndAdd(p2, p3, faceIndex)
			normalizeAndAdd(p3, p0, faceIndex)
			faceIndex++
		}
	}

	// Generate bottom cap faces and edges
	for j := 0; j < sectors; j++ {
		p1 := getIndex(stacks-2, j)
		p2 := getIndex(stacks-2, j+1)
		faces = append(faces, [4]int{bottomPoleIndex, p1, p2, p2})
		normalizeAndAdd(bottomPoleIndex, p1, faceIndex)
		normalizeAndAdd(p1, p2, faceIndex)
		normalizeAndAdd(p2, bottomPoleIndex, faceIndex)
		faceIndex++
	}

	for edge, faceIndices := range edgeToFacesMap {
		edges = append(edges, edge)
		if len(faceIndices) == 2 {
			edgeFaceMap = append(edgeFaceMap, [2]int{faceIndices[0], faceIndices[1]})
		} else {
			edgeFaceMap = append(edgeFaceMap, [2]int{faceIndices[0], faceIndices[0]})
		}
	}

	return vertices, edges, faces, edgeFaceMap
}

func GenerateCylinderData(sectors int, height, radius float32) (
	vertices []mgl32.Vec3,
	edges [][2]int,
	faces [][4]int,
	edgeFaceMap [][2]int,
) {
	// Generate Vertices
	h := height / 2.0
	bottomCenterIndex := 0
	topCenterIndex := 1
	vertices = append(vertices, mgl32.Vec3{0, -h, 0}, mgl32.Vec3{0, h, 0})

	// Add bottom and top ring vertices
	for i := 0; i < sectors; i++ {
		angle := (float64(i) / float64(sectors)) * 2.0 * math.Pi
		x := radius * float32(math.Cos(angle))
		z := radius * float32(math.Sin(angle))
		vertices = append(vertices, mgl32.Vec3{x, -h, z})
	}
	for i := 0; i < sectors; i++ {
		angle := (float64(i) / float64(sectors)) * 2.0 * math.Pi
		x := radius * float32(math.Cos(angle))
		z := radius * float32(math.Sin(angle))
		vertices = append(vertices, mgl32.Vec3{x, h, z})
	}

	getBottomIndex := func(i int) int { return 2 + (i % sectors) }
	getTopIndex := func(i int) int { return 2 + sectors + (i % sectors) }

	// Generate Faces and build an edge-to-face map
	edgeToFacesMap := make(map[[2]int][]int)

	normalizeAndAdd := func(v1, v2, faceIndex int) {
		key := [2]int{v1, v2}
		if v1 > v2 {
			key = [2]int{v2, v1}
		}
		edgeToFacesMap[key] = append(edgeToFacesMap[key], faceIndex)
	}

	faceIndex := 0
	for i := 0; i < sectors; i++ {
		b0 := getBottomIndex(i)
		b1 := getBottomIndex(i + 1)
		t0 := getTopIndex(i)
		t1 := getTopIndex(i + 1)

		// Bottom cap face
		faces = append(faces, [4]int{bottomCenterIndex, b1, b0, b0})
		normalizeAndAdd(bottomCenterIndex, b1, faceIndex)
		normalizeAndAdd(b1, b0, faceIndex)
		normalizeAndAdd(b0, bottomCenterIndex, faceIndex)
		faceIndex++

		// Top cap face
		faces = append(faces, [4]int{topCenterIndex, t0, t1, t1})
		normalizeAndAdd(topCenterIndex, t0, faceIndex)
		normalizeAndAdd(t0, t1, faceIndex)
		normalizeAndAdd(t1, topCenterIndex, faceIndex)
		faceIndex++

		// Side face
		faces = append(faces, [4]int{b0, b1, t1, t0})
		normalizeAndAdd(b0, b1, faceIndex)
		normalizeAndAdd(b1, t1, faceIndex)
		normalizeAndAdd(t1, t0, faceIndex)
		normalizeAndAdd(t0, b0, faceIndex)
		faceIndex++
	}

	for edge, faceIndices := range edgeToFacesMap {
		edges = append(edges, edge)
		if len(faceIndices) == 2 {
			edgeFaceMap = append(edgeFaceMap, [2]int{faceIndices[0], faceIndices[1]})
		} else if len(faceIndices) == 1 {
			edgeFaceMap = append(edgeFaceMap, [2]int{faceIndices[0], faceIndices[0]})
		} else {
			edgeFaceMap = append(edgeFaceMap, [2]int{-1, -1})
		}
	}

	return vertices, edges, faces, edgeFaceMap
}

func GenerateTorusData(mainSegments, tubeSegments int, mainRadius, tubeRadius float32) (
	vertices []mgl32.Vec3,
	edges [][2]int,
	faces [][4]int,
	edgeFaceMap [][2]int,
) {
	// Generate Vertices
	for i := 0; i < mainSegments; i++ {
		mainAngle := (float64(i) / float64(mainSegments)) * 2.0 * math.Pi

		for j := 0; j < tubeSegments; j++ {
			tubeAngle := (float64(j) / float64(tubeSegments)) * 2.0 * math.Pi

			x := (mainRadius + tubeRadius*float32(math.Cos(tubeAngle))) * float32(math.Cos(mainAngle))
			y := (mainRadius + tubeRadius*float32(math.Cos(tubeAngle))) * float32(math.Sin(mainAngle))
			z := tubeRadius * float32(math.Sin(tubeAngle))

			vertices = append(vertices, mgl32.Vec3{x, y, z})
		}
	}

	getIndex := func(i, j int) int {
		return (i%mainSegments)*tubeSegments + (j % tubeSegments)
	}

	// Generate edges and faces
	for i := 0; i < mainSegments; i++ {
		for j := 0; j < tubeSegments; j++ {
			p0 := getIndex(i, j)
			p1 := getIndex(i+1, j)
			p2 := getIndex(i+1, j+1)
			p3 := getIndex(i, j+1)

			currentFaceIndex := i*tubeSegments + j
			neighborFaceIndex_j := i*tubeSegments + (j-1+tubeSegments)%tubeSegments
			neighborFaceIndex_i := (i-1+mainSegments)%mainSegments*tubeSegments + j

			// Add the face
			faces = append(faces, [4]int{p0, p1, p2, p3})

			// Add edges and their corresponding face pairs
			edges = append(edges, [2]int{p0, p1})
			edgeFaceMap = append(edgeFaceMap, [2]int{currentFaceIndex, neighborFaceIndex_j})

			// Edge along tube
			edges = append(edges, [2]int{p0, p3})
			edgeFaceMap = append(edgeFaceMap, [2]int{currentFaceIndex, neighborFaceIndex_i})
		}
	}

	return vertices, edges, faces, edgeFaceMap
}
