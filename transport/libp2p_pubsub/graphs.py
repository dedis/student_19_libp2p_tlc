from matplotlib import pyplot as plt

if __name__ == '__main__':
    Filename = "../../logs/NoFailure_Mail"
    with open(Filename + ".log") as f:
        line = f.readline()
        data = dict()
        while line:
            parts = line.split()
            node, timeStep = parts[2].split(",")
            _, millisec = parts[1].split(".")
            if node not in data.keys():
                times = list()
            else:
                times = data[node]
            times.append((int(timeStep), float(parts[0] + "." + millisec)))
            data[node] = times
            line = f.readline()

    data["0"], data["4"] = data["4"], data["0"]
    data["1"], data["3"] = data["3"], data["1"]

    base = min([a[0][1] for a in list(data.values())])
    max = max([a[-1][1] for a in list(data.values())]) - base
    print(max)
    for j,node in enumerate(data.keys()):
        plt.step([i[1] - base for i in data[node]], [i[0] + 0.03*j for i in data[node]], where='post')

    plt.xlabel("time in seconds")
    plt.ylabel("time steps")
    # plt.savefig("../graphs/" + Filename + ".png")
    plt.show()
    #
    # no = [11,21,31,41,51,61,71,81,91,101]
    # time = [3.18,20,66.9,192,339,793,1355,2125,3080,4029]
    # plt.plot(no,time,'ro',no,time,'r')
    # plt.xlabel("number of nodes")
    # plt.ylabel("execution time for 10 steps")
    # plt.savefig("Test.png")
    # plt.show()

