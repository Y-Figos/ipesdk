local function is_nonempty(str)
    return str ~= nil and str ~= ""
end

local function to_number(val)
    if val == nil or val == "" then
        return nil
    end
    local n = tonumber(val)
    return n
end

function main()
    -- dataframe com TODAS as lojas (lido pelo adapter csv em modo batch)
    local df = Module_1_input

    -- se quiser debugar, pode ver shape:
    -- print("Linhas antes da limpeza:", df:shape())

    local cleaned = df:filter(function(row)
        local produto       = row["produto"]
        local quantidade    = to_number(row["quantidade"])
        local custo_unit    = to_number(row["custo_unitario"])

        -- Produto obrigatório
        if not is_nonempty(produto) then
            return false
        end

        -- Quantidade numérica e não negativa
        if quantidade == nil or quantidade < 0 then
            return false
        end

        -- Custo unitário numérico e não negativo
        if custo_unit == nil or custo_unit < 0 then
            return false
        end

        return true
    end)

    -- print("Linhas depois da limpeza:", cleaned:shape())

    -- o engine vai procurar por "output" na hora de exportar
    return {
        output = cleaned
    }
end
