package model

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"seeccloud.com/edscron/pkg/cronx"
)

var (
	_ AreaModel = (*customAreaModel)(nil)

	munic                        = "市辖区"
	shenzhen                     = "深圳"
	defCities                    = map[string]string{"西安市": "长安区"}
	cacheEdsCronAreaParentPrefix = "cache:edsCron:area:parent:"
	cacheEdsCronAreaRegionPrefix = "cache:edsCron:area:region:"
)

type (
	// AreaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAreaModel.
	AreaModel interface {
		areaModel
		FindParent(ctx context.Context, area string) (*Area, error)
		Get95598Region(ctx context.Context, address string) (*Region, error)
	}

	customAreaModel struct {
		*defaultAreaModel
	}

	Region struct {
		Province string `db:"province"`
		City     string `db:"city"`
		Area     string `db:"area"`
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

func (m *customAreaModel) Get95598Region(ctx context.Context, address string) (*Region, error) {
	var resp Region

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

	// 蒙东电网
	if slices.Contains(cronx.NeimengdongCities, cty) {
		// 省会呼和浩特隶属蒙西电网
		query = fmt.Sprintf("select %s from %s where `parent` = ? and `name` like ? limit 1", areaRows, m.table)
		err = m.QueryRowNoCacheCtx(ctx, &city, query, province.Id, cronx.NeimengdongCities[0]+"%")
		if err != nil {
			return nil, err
		}
	} else {
		// 查省会城市
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

	resp = Region{
		Province: prv,
		City:     city.Name,
		Area:     area.Name,
	}

	m.SetCacheCtx(ctx, edsCronAreaRegionKey, resp)
	return &resp, nil
}
