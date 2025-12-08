-- Module_2: limpa produtos e adiciona faixa_preco

local function faixa_preco_from_valor(v)
    local n = tonumber(v)
    if not n then
        return "desconhecido"
    end
    if n < 100 then
        return "baixo"
    elseif n < 500 then
        return "médio"
    else
        return "alto"
    end
end

function main()
    local df = Module_2_input

    -- 1) remove produtos sem preço ou com preço inválido
    local cleaned = df:filter(function(row)
        local preco = row["preco"]
        if preco == nil then
            return false
        end
        local n = tonumber(preco)
        if n == nil then
            return false
        end
        return n > 0
    end)

    -- 2) adiciona coluna faixa_preco
    local preco_col = cleaned["preco"]
    if preco_col ~= nil then
        local faixa_col = preco_col:apply(function(v)
            return faixa_preco_from_valor(v)
        end)
        cleaned:new_column("faixa_preco", faixa_col)
    end

    return {
        produtos = cleaned
    }
end
