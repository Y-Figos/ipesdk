-- Module_4: consolida visão da operação de vendas

local function is_truthy(v)
    if v == nil then
        return false
    end
    if type(v) == "boolean" then
        return v
    end
    local n = tonumber(v)
    if n ~= nil then
        return n ~= 0
    end
    local s = tostring(v)
    s = string.lower(s)
    return s == "true" or s == "sim"
end

function main()
    -- vindos dos módulos anteriores (limpos/enriquecidos)
    local clientes_df = clientes or Module_1_input
    local produtos_df = produtos or Module_2_input
    local vendas_df   = vendas   or Module_3_input

    -- 1) fat_por_mes: usa o dataframe de vendas enriquecido (com 'mes')
    local fat_por_mes_df = vendas_df

    -- 2) top_produtos: só vendas marcadas como valor_alto
    local top_produtos_df = vendas_df:filter(function(row)
        local v = row["valor_alto"]
        return is_truthy(v)
    end)

    -- 3) top_lojas: exemplo simples filtrando por alguns clientes
    local top_lojas_df = vendas_df:filter(function(row)
        local idc = row["id_cliente"]
        if idc == nil then
            return false
        end
        local n = tonumber(idc)
        if not n then
            return false
        end
        -- aqui você escolhe os "clientes destaque"
        return (n == 10) or (n == 11)
    end)

    return {
        fat_por_mes  = fat_por_mes_df,
        top_produtos = top_produtos_df,
        top_lojas    = top_lojas_df,
    }
end
