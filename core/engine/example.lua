function main()
    reader = new_reader()
    reader:set_path("D:\\Huawei Projects\\1- Project IPE\\Files\\users_100.csv")

    df = reader:get_data()

    filtered = df:filter(function(row)
        return row["age"] < 0 
    end)
    return filtered
end