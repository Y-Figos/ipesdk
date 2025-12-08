-- Module_3: enriquece vendas com 'mes' e 'valor_alto'

local function add_mes_column(df)
    local data_col = df["data"]
    if data_col == nil then
        return df
    end

    local mes_col = data_col:apply(function(v)
        if v == nil then
            return nil
        end
        local s = tostring(v)
        -- assume "YYYY-MM-DD" -> pega "YYYY-MM"
        return string.sub(s, 1, 7)
    end)

    df:new_column("mes", mes_col)
    return df
end

function main()
    local df = Module_3_input

    -- 1) adiciona coluna "mes"
    df = add_mes_column(df)

    -- 2) adiciona coluna booleana "valor_alto" (>= 200)
    local valor_col = df["valor"]
    if valor_col ~= nil then
        local flag_col = valor_col:apply(function(v)
            local n = tonumber(v)
            if not n then
                return false
            end
            return n >= 200
        end)
        df:new_column("valor_alto", flag_col)
    end

    return {
        vendas = df
    }
end
