package types

import (
	"fmt"
	"strings"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeDenomPairTakerFee = "DenomPairTakerFee"
)

func init() {
	govtypesv1.RegisterProposalType(ProposalTypeDenomPairTakerFee)
}

var (
	_ govtypesv1.Content = &DenomPairTakerFeeProposal{}
)

// NewDenomPairTakerFeeProposal returns a new instance of a denom pair taker fee proposal struct.
func NewDenomPairTakerFeeProposal(title, description string, records []DenomPairTakerFee) govtypesv1.Content {
	return &DenomPairTakerFeeProposal{
		Title:             title,
		Description:       description,
		DenomPairTakerFee: records,
	}
}

func (p *DenomPairTakerFeeProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *DenomPairTakerFeeProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *DenomPairTakerFeeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *DenomPairTakerFeeProposal) ProposalType() string {
	return ProposalTypeDenomPairTakerFee
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *DenomPairTakerFeeProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return validateDenomPairTakerFees(p.DenomPairTakerFee)
}

// String returns a string containing the denom pair taker fee proposal.
func (p DenomPairTakerFeeProposal) String() string {
	recordsStr := ""
	for _, record := range p.DenomPairTakerFee {
		recordsStr = recordsStr + fmt.Sprintf("(TokenInDenom: %s, TokenOutDenom: %s, TakerFee: %s) ", record.TokenInDenom, record.TokenOutDenom, record.TakerFee.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Denom Pair Taker Fee Proposal:
Title:       %s
Description: %s
Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
