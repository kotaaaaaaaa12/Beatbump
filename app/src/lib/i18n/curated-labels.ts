type Formatter = (key: string, values?: Record<string, string | number>) => string;
const normalize = (text: string) => text.trim().replace(/[’‘]/g, "'").replace(/\s*&\s*/g, " and ").replace(/\s*\+\s*/g, " and ").replace(/\s+/g, " ").toLowerCase();
const genres: Record<string, string> = {
    pop: "Pop", rock: "Rock", metal: "Metal", jazz: "Jazz", blues: "Blues", country: "Country",
    classical: "Classical", opera: "Opera", soul: "Soul", "neo-soul": "Neo-soul", "neo soul": "Neo-soul",
    "r&b": "R&B", "r and b": "R&B", "adult r and b": "Adult R&B", "hip-hop": "Hip-hop", "hip hop": "Hip-hop",
    rap: "Rap", phonk: "Phonk", "pop-punk": "Pop-punk", "pop punk": "Pop-punk", punk: "Punk",
    "hard rock": "Hard Rock", "classic rock": "Classic Rock", "alternative metal": "Alternative Metal",
    indie: "Indie", alternative: "Alternative", "indie rock": "Indie Rock", "alternative indie": "Indie and Alternative",
    "indie and alternative": "Indie and Alternative", "alt-pop": "Alternative Pop", "indie pop": "Indie Pop",
    acoustic: "Acoustic", folk: "Folk", "folk and acoustic": "Folk and Acoustic", reggae: "Reggae",
    latin: "Latin", "latin pop": "Latin Pop", "latin hip hop": "Latin Hip-hop", reggaeton: "Reggaeton",
    "j-pop": "J-Pop", "jpop": "J-Pop", "k-pop": "K-Pop", "kpop": "K-Pop", "j-rap": "J-Rap",
    dance: "Dance", "dance pop": "Dance Pop", "dance-pop": "Dance Pop", electronic: "Electronic", edm: "EDM",
    techno: "Techno", house: "House", "house music": "House", "progressive house": "Progressive House",
    "deep house": "Deep House", trance: "Trance", "drum and bass": "Drum and Bass", disco: "Disco", funk: "Funk",
    "slow jams": "Slow Jams", lofi: "Lo-fi", "lo-fi": "Lo-fi", "ambient": "Ambient", gospel: "Gospel",
    bollywood: "Bollywood", soundtracks: "Soundtracks", musicals: "Musicals", christmas: "Christmas Music",
};
const regions: Record<string, string> = {
    japanese: "Region JP", japan: "Region JP", korean: "Region KR", french: "Region FR", german: "Region DE",
    brazilian: "Region BR", turkish: "Region TR", ukrainian: "Region UA", indian: "Region IN", uk: "Region GB",
    american: "Region US", spanish: "Region ES", italian: "Region IT", vietnamese: "Region VN", thai: "Region TH",
    chinese: "Region CN", "south african": "Region ZA", indonesian: "Region ID", filipino: "Region PH", opm: "Region PH",
};
const modifiers: Record<string, string> = {
    new: "New {genre}", mellow: "Mellow {genre}", essential: "Essential {genre}", ultimate: "Ultimate {genre}",
    happy: "Happy {genre}", "feel-good": "Feel-good {genre}", "feel good": "Feel-good {genre}",
    chill: "Chill {genre}", chilled: "Chill {genre}", relaxing: "Relaxing {genre}", romantic: "Romantic {genre}",
    summer: "Summer {genre}", "laid-back": "Laid-back {genre}", "laid back": "Laid-back {genre}",
};
const suffixes: Record<string, string> = {
    hits: "{genre} hits", classics: "{genre} classics", essentials: "Essential {genre}",
    anthems: "{genre} anthems", workout: "{genre} workout", party: "{genre} party", "party hits": "{genre} party hits",
    chill: "Chill {genre}", hotlist: "{genre} hotlist", instrumentals: "{genre} instrumentals", love: "{genre} love songs",
};
const own = (map: Record<string, string>, key: string) => Object.prototype.hasOwnProperty.call(map, key) ? map[key] : undefined;

function genreLabel(text: string, translate: Formatter): string | undefined {
    const normalized = normalize(text);
    const direct = own(genres, normalized);
    if (direct) return translate(direct);
    for (const [name, key] of Object.entries(regions)) {
        if (!normalized.startsWith(name + " ")) continue;
        const genre = own(genres, normalized.slice(name.length + 1));
        if (genre) return translate("{region} {genre}", { region: translate(key), genre: translate(genre) });
    }
    return undefined;
}

function collectionLabel(text: string, translate: Formatter): string | undefined {
    const direct = genreLabel(text, translate);
    if (direct) return direct;
    const normalized = normalize(text);
    for (const [suffix, key] of Object.entries(suffixes).sort(([a], [b]) => b.length - a.length)) {
        if (!normalized.endsWith(" " + suffix)) continue;
        const genre = genreLabel(normalized.slice(0, -(suffix.length + 1)), translate);
        if (genre) return translate(key, { genre });
    }
    for (const [modifier, key] of Object.entries(modifiers)) {
        if (!normalized.startsWith(modifier + " ")) continue;
        const genre = genreLabel(normalized.slice(modifier.length + 1), translate);
        if (genre) return translate(key, { genre });
    }
    return undefined;
}

// This grammar is limited to discovery labels. Ordinary track, artist and
// user-playlist names never pass through it.
export function translateCuratedLabel(text: string, translate: Formatter): string | undefined {
    const present = /^presenting (.+)$/i.exec(text.trim());
    if (present) return translate("Presenting {artist}", { artist: present[1] });
    const normalized = normalize(text);
    const direct = collectionLabel(normalized, translate);
    if (direct) return direct;
    const year = /^(.+) (19\d{2}|20\d{2})$/.exec(normalized);
    if (year) {
        const collection = collectionLabel(year[1], translate);
        if (collection) return translate("{year}: {collection}", { year: year[2], collection });
    }
    const decade = /^'?(\d{2}|(?:19|20)\d{2})s (.+)$/.exec(normalized);
    if (decade) {
        const collection = collectionLabel(decade[2], translate);
        if (collection) {
            const number = Number(decade[1]);
            const period = decade[1].length === 4 ? number : number < 30 ? 2000 + number : 1900 + number;
            return translate("{period}s {collection}", { period, collection });
        }
    }
    const feelGood = /^feelin' good in (?:the )?'?(\d{2})s$/.exec(normalized);
    if (feelGood) {
        const number = Number(feelGood[1]);
        return translate("Feel-good songs from the {period}s", { period: number < 30 ? 2000 + number : 1900 + number });
    }
    return undefined;
}
