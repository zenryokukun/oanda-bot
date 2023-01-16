import datetime
import matplotlib.pyplot as plt
import json
import sys

BALANCE_F = "./balance.json"
TRADE_F = "./trade.json"


def load(fpath: str) -> dict:
    data = None
    with open(fpath, mode="r") as f:
        data = json.load(f)
    return data


def time_str(jobj: dict):
    jobj["time"] = [datetime.datetime.fromtimestamp(v) for v in jobj["X"]]


def slice_trade(start_time, tobj):
    """tobj["time"]の、start_timeより前の部分を落とす

    Args:
        start_time (datetime): balance["time"]の最初の値を想定
        tobj (dict): trade.jsonをloadしたもの
    """
    i = 0
    for i, t in enumerate(tobj["time"]):
        if t > start_time:
            break

    tobj["time"] = tobj["time"][i:]
    tobj["X"] = tobj["X"][i:]
    tobj["Y"] = tobj["Y"][i:]
    tobj["Action"] = tobj["Action"][i:]
    tobj["Side"] = tobj["Side"][i:]

    return tobj


def graph(img_path: str):
    # ファイルから読み取る
    bl = load(BALANCE_F)
    tr = load(TRADE_F)
    # unix時間を文字列に変換した値をセット
    time_str(bl)
    time_str(tr)
    # tradeがグラフからはみ出るので、最初のbl["time"]より前のデータは落とす。
    tr = slice_trade(bl["time"][0], tr)
    # trデータを分割
    openbuy_x = []
    openbuy_y = []
    opensell_x = []
    opensell_y = []
    close_x = []
    close_y = []
    for i in range(len(tr["X"])):
        _x = tr["time"][i]
        _y = tr["Y"][i]
        if tr["Action"][i] == "OPEN":
            if tr["Side"][i] == "BUY":
                openbuy_x.append(_x)
                openbuy_y.append(_y)
            else:
                opensell_x.append(_x)
                opensell_y.append(_y)
        else:
            close_x.append(_x)
            close_y.append(_y)

    # グラフ
    fig = plt.figure()
    # 左グラフ 実際の価格
    ax = fig.add_subplot(111)
    ax.set_ylabel("USD_JPY")
    ax.plot(bl["time"], bl["Y"], label="USD/JPY", color="orange")
    # 取引箇所
    ax.scatter(openbuy_x, openbuy_y, label="@openBuy", color="red")
    ax.scatter(opensell_x, opensell_y,
               label="@openSell", color="lime")
    ax.scatter(close_x, close_y, label="@close",
               facecolors="none", edgecolors="black", s=80)

    # 右グラフ
    # 利益推移
    ax2 = ax.twinx()
    ax2.set_ylabel("Profit/Loss")
    ax2.plot(bl["time"], bl["TotalPL"], label="TotalPL")

    plt.title("Oanda Trade Result")
    plt.xlabel("TIME")
    ax.legend(loc=2)
    ax2.legend(loc=3)
    plt.gcf().autofmt_xdate()
    plt.tight_layout()
    plt.grid(True)
    # plt.show()
    plt.savefig(img_path)


if __name__ == "__main__":
    try:
        tpath = sys.argv[1]
    except IndexError as err:
        print(err)
        sys.exit()
    graph(tpath)
