package model

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/cronx"
)

var (
	_                             AreaModel = (*customAreaModel)(nil)
	munic                                   = "市辖区"
	defCities                               = map[string]string{"西安市": "长安区"}
	cacheEdsCronAreaParentPrefix            = "cache:edsCron:area:parent:"
	cacheEdsCronAreaRegionPrefix            = "cache:edsCron:area:region:"
	cacheEdsCronAreaCapitalPrefix           = "cache:edsCron:area:capital:"
)

type (
	// AreaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAreaModel.
	AreaModel interface {
		areaModel
		FindParent(ctx context.Context, area string) (*Area, error)
		Get95598Address(ctx context.Context, address string) (*Address, error)
		GetProvincialCapital(ctx context.Context, area string) (string, error)
	}

	customAreaModel struct {
		*defaultAreaModel
	}
)

// NewAreaModel returns a model for the database table.
func NewAreaModel(conn sqlx.SqlConn, c cache.CacheConf) AreaModel {
	return &customAreaModel{
		defaultAreaModel: newAreaModel(conn, c),
	}
}

func (m *customAreaModel) FindParent(ctx context.Context, area string) (*Area, error) {
	edsCronAreaAddressKey := fmt.Sprintf("%s%v", cacheEdsCronAreaParentPrefix, area)
	var resp, parent Area
	err := m.GetCacheCtx(ctx, edsCronAreaAddressKey, &parent)
	if err == nil {
		if len(parent.Name) == 0 {
			return nil, ErrNotFound
		}
		return &parent, nil
	}

	err = m.QueryRowCtx(ctx, &resp, edsCronAreaAddressKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `name` like ? limit 1", areaRows, m.table)
		err := conn.QueryRowCtx(ctx, &resp, query, "%"+area+"%")
		if err != nil {
			return err
		}

		if resp.Parent == -1 {
			return ErrNotFound
		}

		one, err := m.FindOne(ctx, uint64(resp.Parent))
		if err != nil {
			return err
		}

		parent = *one
		return nil
	})

	switch err {
	case nil:
		m.SetCacheCtx(ctx, edsCronAreaAddressKey, parent)
		return &parent, nil
	case sqlc.ErrNotFound:
		m.SetCacheCtx(ctx, edsCronAreaAddressKey, nil)
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAreaModel) Get95598Address(ctx context.Context, address string) (*Address, error) {
	var resp Address

	prv, cty := cronx.ExtractAddress(address, true)
	resp.Province = prv
	resp.City = cty

	// 南方电网
	if _, ok := cronx.NanfangfProvinces[prv]; ok {
		return &resp, nil
	}

	edsCronAreaRegionKey := fmt.Sprintf("%s%v", cacheEdsCronAreaRegionPrefix, address)
	err := m.GetCacheCtx(ctx, edsCronAreaRegionKey, &resp)
	if err == nil {
		if len(resp.Province) == 0 {
			return nil, ErrNotFound
		}
		return &resp, nil
	}

	var province, city, area Area
	query := fmt.Sprintf("select %s from %s where `name` like ? limit 1", areaRows, m.table)
	err = m.QueryRowNoCacheCtx(ctx, &province, query, prv+"%")
	if err != nil {
		return nil, err
	}

	// 相同电价区域应用默认城市，避免出现重复任务，如"厦门"使用"福州"
	if defCity := cronx.GetDefaultCity(cty); len(defCity) > 0 {
		// 非省会城市
		query = fmt.Sprintf("select %s from %s where `parent` = ? and `name` like ? limit 1", areaRows, m.table)
		err = m.QueryRowNoCacheCtx(ctx, &city, query, province.Id, defCity+"%")
		if err != nil {
			return nil, err
		}
	} else {
		// 其他默认省会城市
		query = fmt.Sprintf("select %s from %s where `parent` = ? order by `id` limit 1", areaRows, m.table)
		err = m.QueryRowNoCacheCtx(ctx, &city, query, province.Id)
		if err != nil {
			return nil, err
		}
	}

	query = fmt.Sprintf("select %s from %s where `parent` = ? order by `id` limit 1", areaRows, m.table)
	err = m.QueryRowNoCacheCtx(ctx, &area, query, city.Id)
	if err != nil {
		return nil, err
	}

	// 直辖市
	if strings.HasPrefix(city.Name, province.Name) {
		city.Name = munic
	}

	// 西安市：在95598.cn从"西安市咸阳地区"而非"西安市"选择默认区县
	if ar, ok := defCities[city.Name]; ok {
		area.Name = ar
	}

	resp = Address{
		Province: prv,
		City:     city.Name,
		Area:     area.Name,
	}

	m.SetCacheCtx(ctx, edsCronAreaRegionKey, resp)
	return &resp, nil
}

func (m *customAreaModel) GetProvincialCapital(ctx context.Context, area string) (string, error) {
	key := fmt.Sprintf("%s%v", cacheEdsCronAreaCapitalPrefix, area)
	capital := ""
	err := m.GetCacheCtx(ctx, key, &capital)
	if err == nil {
		if len(capital) == 0 {
			return "", ErrNotFound
		}

		return capital, nil
	}

	area = regexp.MustCompile(`\p{Han}+`).FindString(area)
	var one Area
	query := fmt.Sprintf("select %s from %s where `name` like ? limit 1", areaRows, m.table)
	err = m.QueryRowNoCacheCtx(ctx, &one, query, area+"%")
	if err != nil {
		return "", err
	}

	for one.Parent > 0 {
		parent, err := m.FindParent(ctx, one.Name)
		if err != nil {
			return "", err
		}

		one = *parent
	}

	query = fmt.Sprintf("select %s from %s where `parent` = ? limit 1", areaRows, m.table)
	err = m.QueryRowNoCacheCtx(ctx, &one, query, one.Id)
	if err != nil {
		return "", err
	}

	capital = cronx.Shorten(one.Name)
	m.SetCacheCtx(ctx, key, capital)
	return capital, nil
}
