package bitbox

type NullParams struct{}

type ParamSet map[string]string

func newParamsForType(cellType string) (any, error) {
	switch cellType {
	case "sample":
		return &SampleParams{}, nil
	case "samtempl":
		return &SamTemplateParams{}, nil
	case "delay":
		return &DelayParams{}, nil
	case "reverb":
		return &ReverbParams{}, nil
	case "filter":
		return &FilterParams{}, nil
	case "bitcrusher":
		return &BitcrusherParams{}, nil
	case "ioconnectin":
		return &IOConnectInParams{}, nil
	case "ioconnectout":
		return &IOConnectOutParams{}, nil
	case "song":
		return &SongParams{}, nil
	case "noteseq":
		return &NoteseqParams{}, nil
	case "asset":
		return &AssetParams{}, nil
	case "section":
		return &SectionParams{}, nil
	case "eq":
		return &EqParams{}, nil
	case "null":
		return &NullParams{}, nil
	default:
		// Fallback to empty params.
		ps := ParamSet{}
		return &ps, nil
	}
}
