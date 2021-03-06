package chain

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"go.uber.org/zap"
)

// GetBlockHeaderHash get block header for verify
func GetBlockHeaderHash(block *BlockHeader) Checksum256 {
	raws, _ := MarshalBinary(block)

	h := sha256.New()
	_, _ = h.Write(raws)

	return h.Sum(nil)
}

// HashCheckSumPair get sha256 hash from c1+c2
func HashCheckSumPair(c1, c2 Checksum256) Checksum256 {
	h := sha256.New()

	if len(c1) == 0 {
		h.Write(bytes.Repeat([]byte{0}, TypeSize.Checksum256))
	} else {
		h.Write(c1)
	}

	if len(c2) == 0 {
		h.Write(bytes.Repeat([]byte{0}, TypeSize.Checksum256))
	} else {
		h.Write(c2)
	}

	return h.Sum(nil)
}

// IsSamePubKey p1 == p2
func IsSamePubKey(p1, p2 PublicKey) bool {
	return p1.Curve == p2.Curve && bytes.Equal(p1.Content, p2.Content)
}

// getSigDigest get sig digest for verifier
func (c *Chain) getSigDigest(block *SignedBlock) (Checksum256, error) {
	headerHash := GetBlockHeaderHash(&block.BlockHeader)
	scheduleProducersHash := c.ScheduleProducers.GetScheduleProducersHash()
	blockrootMerkle := c.PendingState.BlockrootMerkle.GetRoot()

	headerAndBmroot := HashCheckSumPair(headerHash, blockrootMerkle)
	sigDigest := HashCheckSumPair(headerAndBmroot, scheduleProducersHash)

	/*
		v.logger.Debug("sigDigest info",
			zap.String("headerHash", headerHash.String()),
			zap.String("scheduleProducersHash", scheduleProducersHash.String()),
			zap.String("headerAndBmroot", headerAndBmroot.String()),
			zap.String("sigDigest", sigDigest.String()))
	*/
	return sigDigest, nil
}

// verifySign verify if block is sign right
func (c *Chain) verifySign(block *SignedBlock) error {
	sigDigest, err := c.getSigDigest(block)
	if err != nil {
		return err
	}

	pubKey, err := c.ScheduleProducers.GetScheduleProducer(block.ScheduleVersion, block.Producer)
	if err != nil {
		return err
	}

	signPubKey, err := block.ProducerSignature.PublicKey(sigDigest)
	if err != nil {
		return err
	}

	if !IsSamePubKey(pubKey.BlockSigningKey, signPubKey) {
		c.logger.Error("sign err",
			zap.Uint32("blockNum", block.BlockNumber()),
			zap.String("pubkey", pubKey.BlockSigningKey.String()),
			zap.String("signPubKey", signPubKey.String()),
			zap.Uint32("scheduleVersion", block.ScheduleVersion),
			zap.String("producer", string(block.Producer)),
		)
		return errors.New("sign error")
	}

	return nil
}
