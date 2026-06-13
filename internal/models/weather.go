package models

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	WeatherCycle   = 2160
	VanaBaseEpoch  = int64(1009810800)
	VanaDaySeconds = int64(57*60 + 36)
)

type Weather uint8

const (
	WEATHER_NONE          Weather = 0
	WEATHER_SUNSHINE      Weather = 1
	WEATHER_CLOUDS        Weather = 2
	WEATHER_FOG           Weather = 3
	WEATHER_HOT_SPELL     Weather = 4
	WEATHER_HEAT_WAVE     Weather = 5
	WEATHER_RAIN          Weather = 6
	WEATHER_SQUALL        Weather = 7
	WEATHER_DUST_STORM    Weather = 8
	WEATHER_SAND_STORM    Weather = 9
	WEATHER_WIND          Weather = 10
	WEATHER_GALES         Weather = 11
	WEATHER_SNOW          Weather = 12
	WEATHER_BLIZZARDS     Weather = 13
	WEATHER_THUNDER       Weather = 14
	WEATHER_THUNDERSTORMS Weather = 15
	WEATHER_AURORAS       Weather = 16
	WEATHER_STELLAR_GLARE Weather = 17
	WEATHER_GLOOM         Weather = 18
	WEATHER_DARKNESS      Weather = 19
)

func (w Weather) String() string {
	names := map[Weather]string{
		WEATHER_NONE:          "NONE",
		WEATHER_SUNSHINE:      "SUNSHINE",
		WEATHER_CLOUDS:        "CLOUDS",
		WEATHER_FOG:           "FOG",
		WEATHER_HOT_SPELL:     "HOT_SPELL",
		WEATHER_HEAT_WAVE:     "HEAT_WAVE",
		WEATHER_RAIN:          "RAIN",
		WEATHER_SQUALL:        "SQUALL",
		WEATHER_DUST_STORM:    "DUST_STORM",
		WEATHER_SAND_STORM:    "SAND_STORM",
		WEATHER_WIND:          "WIND",
		WEATHER_GALES:         "GALES",
		WEATHER_SNOW:          "SNOW",
		WEATHER_BLIZZARDS:     "BLIZZARDS",
		WEATHER_THUNDER:       "THUNDER",
		WEATHER_THUNDERSTORMS: "THUNDERSTORMS",
		WEATHER_AURORAS:       "AURORAS",
		WEATHER_STELLAR_GLARE: "STELLAR_GLARE",
		WEATHER_GLOOM:         "GLOOM",
		WEATHER_DARKNESS:      "DARKNESS",
	}
	if s, ok := names[w]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN_%d", w)
}

type DayWeather struct {
	VanaDay int64
	Normal  Weather
	Common  Weather
	Rare    Weather
}

type WeatherForecast struct {
	Day     int
	Weather DayWeather
}

type WeatherBlob struct {
	packed []uint16
}

func NewWeatherBlob(sqlHex string) (*WeatherBlob, error) {
	sqlHex = strings.TrimSpace(strings.TrimPrefix(sqlHex, "0x"))
	raw, err := hex.DecodeString(sqlHex)
	if err != nil {
		return nil, err
	}
	if len(raw)%2 != 0 {
		return nil, fmt.Errorf("odd blob length: %d", len(raw))
	}

	packed := make([]uint16, len(raw)/2)
	for i := range len(packed) {
		packed[i] = binary.LittleEndian.Uint16(raw[i*2 : i*2+2])
	}

	return &WeatherBlob{packed: packed}, nil
}

func DecodePacked(v uint16) (Weather, Weather, Weather) {
	return Weather(v >> 10), Weather((v >> 5) & 0x1F), Weather(v & 0x1F)
}

func weatherCycleIndex(vanaDay int64) int {
	idx := int(vanaDay % WeatherCycle)
	if idx < 0 {
		idx += WeatherCycle
	}
	return idx
}

func currentVanaDayFromUnix(unix int64) int64 {
	return (unix - VanaBaseEpoch) / VanaDaySeconds
}

func (w *WeatherBlob) WeatherForVanaDay(vanaDay int64) (DayWeather, error) {
	if len(w.packed) < WeatherCycle {
		return DayWeather{}, fmt.Errorf("expected at least %d entries, got %d", WeatherCycle, len(w.packed))
	}

	idx := weatherCycleIndex(vanaDay)
	n, c, r := DecodePacked(w.packed[idx])

	return DayWeather{
		VanaDay: vanaDay,
		Normal:  n,
		Common:  c,
		Rare:    r,
	}, nil
}

func (w *WeatherBlob) WeatherForNowFallback(now time.Time) (DayWeather, error) {
	return w.WeatherForVanaDay(currentVanaDayFromUnix(now.Unix()))
}

func BuildWeatherSequence(seq [][3]uint8) []byte {
	out := make([]byte, 0, len(seq)*2)
	for _, t := range seq {
		v := uint16(t[0])<<10 | uint16(t[1]&0x1F)<<5 | uint16(t[2]&0x1F)
		out = append(out, byte(v>>8), byte(v))
	}
	return out
}

func (w *WeatherBlob) BuildWeatherForecast(days int, now time.Time) ([]WeatherForecast, error) {
	var fullForecast []WeatherForecast
	for d := range days {
		weather, err := w.WeatherForVanaDay(currentVanaDayFromUnix(now.Unix()) + int64(d))
		if err != nil {
			return []WeatherForecast{}, nil
		}

		forecast := WeatherForecast{
			Day:     d,
			Weather: weather,
		}

		fullForecast = append(fullForecast, forecast)
	}
	return fullForecast, nil
}

func FormatWeatherForcastPlain(forecast []WeatherForecast) string {
	var sb strings.Builder
	sb.WriteString(
		fmt.Sprintf(
			"Day | Normal | Common | Rare\n",
		),
	)
	for f := range len(forecast) {
		report := forecast[f]
		sb.WriteString(
			fmt.Sprintf(
				"%v | %v | %v | %v\n",
				report.Day,
				report.Weather.Normal,
				report.Weather.Common,
				report.Weather.Rare,
			),
		)
	}
	return sb.String()
}

func ReportForecast(forecast []WeatherForecast) error {
	for f := range len(forecast) {
		report := forecast[f]
		log.Printf(
			"Day: %v, VandaDay: %v, Normal: %v | Common: %v | Rare: %v",
			// "Day: %v, VandaDay: %v, Common: %v | Normal: %v | Rare: %v",
			report.Day,
			report.Weather.VanaDay,
			report.Weather.Normal,
			report.Weather.Common,
			// report.Weather.Normal,
			report.Weather.Rare,
		)
	}
	return nil
}
