package data_access

import (
	"dating-backend/internal/models"
	"errors"
	"math"
	"sort"
)

func GetRecommendations(userID int64, limit int, maxDistanceKm float64) ([]models.User, error) {
    // сначала получаем текущего пользователя, чтобы знать его coords
    me, err := GetUserByID(userID)
    if err != nil {
        return nil, err
    }

    if me.Latitude == nil || me.Longitude == nil {
        return nil, errors.New("current user has no coordinates")
    }
    lat1 := *me.Latitude
    lon1 := *me.Longitude

    // выбираем пользователей, которых он ещё не свайпнул
    rows, err := DB.Query(`
        SELECT id, username, name, gender, birthday, interested_in, bio, photo_url, latitude, longitude
        FROM users
        WHERE id != ? 
          AND id NOT IN (SELECT target_id FROM swipes WHERE user_id = ?)
    `, userID, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var recs []models.User
    for rows.Next() {
        var u models.User
        err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Gender, &u.Birthday, &u.InterestedIn, &u.Bio, &u.PhotoURL, &u.Latitude, &u.Longitude)
        if err != nil {
            return nil, err
        }
        // пропускаем пользователей без координат
        if u.Latitude == nil || u.Longitude == nil {
            continue
        }
        // вычисляем расстояние
        dist := haversine(lat1, lon1, *u.Latitude, *u.Longitude)
        if dist <= maxDistanceKm {
            recs = append(recs, u)
        }
    }

    // можно отсортировать по расстоянию
    sort.Slice(recs, func(i, j int) bool {
        di := haversine(lat1, lon1, *recs[i].Latitude, *recs[i].Longitude)
        dj := haversine(lat1, lon1, *recs[j].Latitude, *recs[j].Longitude)
        return di < dj
    })

    if len(recs) > limit {
        recs = recs[:limit]
    }
    return recs, nil
}

// расстояние в километрах между двумя точками (lat1, lon1) и (lat2, lon2)
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371 // радиус Земли в км
    dLat := (lat2 - lat1) * math.Pi / 180.0
    dLon := (lon2 - lon1) * math.Pi / 180.0
    lat1Rad := lat1 * math.Pi / 180.0
    lat2Rad := lat2 * math.Pi / 180.0

    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c
}