package crpc

import (
	"fmt"
	"github.com/bitlum/connector/connectors"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func convertProtoMessage(resp proto.Message) string {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
		OrigName:     true,
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		return fmt.Sprintf("unable to decode response: %v", err)
	}

	return jsonStr
}

func convertPaymentStatusToProto(status connectors.PaymentStatus) (PaymentStatus, error) {
	var protoStatus PaymentStatus
	switch status {
	case connectors.Waiting:
		protoStatus = PaymentStatus_WAITING
	case connectors.Completed:
		protoStatus = PaymentStatus_COMPLETED
	case connectors.Pending:
		protoStatus = PaymentStatus_PENDING
	case connectors.Failed:
		protoStatus = PaymentStatus_FAILED
	default:
		return protoStatus, errors.Errorf("unable convert unknown status: %v",
			status)
	}

	return protoStatus, nil
}

func convertAssetToProto(asset connectors.Asset) (Asset, error) {
	var protoAsset Asset
	switch asset {
	case connectors.BTC:
		protoAsset = Asset_BTC
	case connectors.BCH:
		protoAsset = Asset_BCH
	case connectors.ETH:
		protoAsset = Asset_ETH
	case connectors.LTC:
		protoAsset = Asset_LTC
	case connectors.DASH:
		protoAsset = Asset_DASH
	default:
		return protoAsset, errors.Errorf("unable convert unknown asset: %v",
			asset)
	}

	return protoAsset, nil
}

func convertPaymentDirectionToProto(direction connectors.PaymentDirection) (PaymentDirection,
	error) {
	var protoDirection PaymentDirection
	switch direction {
	case connectors.Outgoing:
		protoDirection = PaymentDirection_OUTGOING
	case connectors.Incoming:
		protoDirection = PaymentDirection_INCOMING
	default:
		return protoDirection, errors.Errorf("unable convert unknown direction: %v",
			direction)
	}

	return protoDirection, nil
}

func convertPaymentSystemToProto(system connectors.PaymentSystem) (
	PaymentSystem, error) {
	var protoSystem PaymentSystem
	switch system {
	case connectors.Internal:
		protoSystem = PaymentSystem_INTERNAL
	case connectors.External:
		protoSystem = PaymentSystem_EXTERNAL
	default:
		return protoSystem, errors.Errorf("unable convert unknown system: %v",
			system)
	}

	return protoSystem, nil
}

func convertMediaToProto(media connectors.PaymentMedia) (Media, error) {
	var protoMedia Media
	switch media {
	case connectors.Blockchain:
		protoMedia = Media_BLOCKCHAIN
	case connectors.Lightning:
		protoMedia = Media_LIGHTNING
	default:
		return protoMedia, errors.Errorf("unable convert unknown media: %v",
			media)
	}

	return protoMedia, nil
}

func convertPaymentToProto(payment *connectors.Payment) (*Payment, error) {
	status, err := convertPaymentStatusToProto(payment.Status)
	if err != nil {
		return nil, err
	}

	direction, err := convertPaymentDirectionToProto(payment.Direction)
	if err != nil {
		return nil, err
	}

	asset, err := convertAssetToProto(payment.Asset)
	if err != nil {
		return nil, err
	}

	media, err := convertMediaToProto(payment.Media)
	if err != nil {
		return nil, err
	}

	system, err := convertPaymentSystemToProto(payment.System)
	if err != nil {
		return nil, err
	}

	return &Payment{
		PaymentId: payment.PaymentID,
		UpdatedAt: payment.UpdatedAt,
		Status:    status,
		Direction: direction,
		System:    system,
		Asset:     asset,
		Media:     media,
		Receipt:   payment.Receipt,
		Amount:    payment.Amount.String(),
		MediaFee:  payment.MediaFee.String(),
		MediaId:   payment.MediaID,
	}, nil
}

func ConvertPaymentStatusFromProto(protoStatus PaymentStatus) (
	connectors.PaymentStatus, error) {
	var status connectors.PaymentStatus
	switch protoStatus {
	case PaymentStatus_WAITING:
		status = connectors.Waiting
	case PaymentStatus_COMPLETED:
		status = connectors.Completed
	case PaymentStatus_PENDING:
		status = connectors.Pending
	case PaymentStatus_FAILED:
		status = connectors.Failed
	default:
		return status, errors.Errorf("unable convert unknown status: %v",
			protoStatus)
	}

	return status, nil
}

func ConvertAssetFromProto(protoAsset Asset) (connectors.Asset, error) {
	var asset connectors.Asset
	switch protoAsset {
	case Asset_BTC:
		asset = connectors.BTC
	case Asset_BCH:
		asset = connectors.BCH
	case Asset_ETH:
		asset = connectors.ETH
	case Asset_LTC:
		asset = connectors.LTC
	case Asset_DASH:
		asset = connectors.DASH
	default:
		return asset, errors.Errorf("unable convert unknown asset: %v",
			protoAsset)
	}

	return asset, nil
}

func ConvertPaymentDirectionFromProto(protoDirection PaymentDirection) (
	connectors.PaymentDirection, error) {
	var direction connectors.PaymentDirection
	switch protoDirection {
	case PaymentDirection_OUTGOING:
		direction = connectors.Outgoing
	case PaymentDirection_INCOMING:
		direction = connectors.Incoming
	default:
		return direction, errors.Errorf("unable convert unknown direction: %v",
			protoDirection)
	}

	return direction, nil
}

func ConvertPaymentSystemFromProto(protoSystem PaymentSystem) (
	connectors.PaymentSystem, error) {
	var system connectors.PaymentSystem
	switch protoSystem {
	case PaymentSystem_INTERNAL:
		system = connectors.Internal
	case PaymentSystem_EXTERNAL:
		system = connectors.External
	default:
		return system, errors.Errorf("unable convert unknown system: %v",
			protoSystem)
	}

	return system, nil
}

func ConvertMediaFromProto(protoMedia Media) (connectors.PaymentMedia, error) {
	var media connectors.PaymentMedia
	switch protoMedia {
	case Media_BLOCKCHAIN:
		media = connectors.Blockchain
	case Media_LIGHTNING:
		media = connectors.Lightning
	default:
		return media, errors.Errorf("unable convert unknown media: %v",
			protoMedia)
	}

	return media, nil
}
