export interface Meetup {
  id: string;
  title: string;
  game: string;
  host_name: string;
  host_uid: string;
  lat: number;
  lng: number;
  players_count: number;
  players_needed: number;
  time: string;
  color: string;
}

export class MeetupStore {
  private meetups: Meetup[] = [
    {
      id: "1",
      title: "Hội Ma Sói Đêm Trăng Q1",
      game: "Ultimate Werewolf",
      host_name: "Minh Tuấn",
      host_uid: "default-host-1",
      lat: 10.7769,
      lng: 106.7009,
      players_count: 11,
      players_needed: 15,
      time: "2026-07-10T19:30",
      color: "#bca0f5",
    },
    {
      id: "2",
      title: "Sân Chơi Mèo Nổ Q3",
      game: "Exploding Kittens",
      host_name: "Thanh Trúc",
      host_uid: "default-host-2",
      lat: 10.7828,
      lng: 106.6896,
      players_count: 4,
      players_needed: 5,
      time: "2026-07-11T15:00",
      color: "#ffa4b2",
    },
    {
      id: "3",
      title: "CLB Cờ Tỷ Phú Bình Thạnh",
      game: "Monopoly Deal",
      host_name: "Khánh Huy",
      host_uid: "default-host-3",
      lat: 10.7981,
      lng: 106.7051,
      players_count: 3,
      players_needed: 6,
      time: "2026-07-12T18:00",
      color: "#ffe869",
    },
    {
      id: "4",
      title: "Chiến Thần Catan Hoàn Kiếm",
      game: "Settlers of Catan",
      host_name: "Hoàng Lâm",
      host_uid: "default-host-4",
      lat: 21.0285,
      lng: 105.8542,
      players_count: 2,
      players_needed: 4,
      time: "2026-07-11T19:00",
      color: "#9ee3b2",
    },
    {
      id: "5",
      title: "Hội Avalon Tây Hồ",
      game: "Avalon",
      host_name: "Thu Giang",
      host_uid: "default-host-5",
      lat: 21.0588,
      lng: 105.8285,
      players_count: 5,
      players_needed: 10,
      time: "2026-07-12T14:30",
      color: "#a4f0fd",
    },
  ];

  public getAll(): Meetup[] {
    return [...this.meetups];
  }

  public add(meetup: Meetup): void {
    this.meetups.push(meetup);
  }
}

export const meetupStore = new MeetupStore();
