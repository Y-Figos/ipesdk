-- Module_1: limpa clientes e adiciona faixa_etaria

local function is_nonempty(str)
    return str ~= nil and str ~= ""
end

local function contains_at(str)
    if str == nil then return false end
    return string.find(str, "@", 1, true) ~= nil
end

local function parse_age(raw)
    if raw == nil or raw == "" then
        return nil
    end
    local n = tonumber(raw)
    if n == nil then
        return nil
    end
    return n
end

local function is_valid_age(raw)
    local n = parse_age(raw)
    if n == nil then
        return false
    end
    return n >= 0 and n <= 120
end

local function is_valid_address(addr)
    if not is_nonempty(addr) then
        return false
    end
    local lower = string.lower(addr)
    if lower == "nao identificado" then
        return false
    end
    return true
end

local function faixa_etaria_from_idade(v)
    local n = tonumber(v)
    if not n then
        return "desconhecida"
    end
    if n < 18 then
        return "menor"
    elseif n <= 30 then
        return "jovem adulto"
    elseif n <= 50 then
        return "adulto"
    else
        return "sênior"
    end
end

function main()
    -- dataframe injetado pelo engine
    local df = Module_1_input

    -- 1) filtra linhas inválidas
    local cleaned = df:filter(function(row)
        local nome     = row["nome"]
        local idade    = row["idade"]
        local email    = row["email"]
        local endereco = row["endereco"]

        if not is_nonempty(nome) then
            return false
        end

        if not is_nonempty(email) or not contains_at(email) then
            return false
        end

        if not is_valid_age(idade) then
            return false
        end

        if not is_valid_address(endereco) then
            return false
        end

        return true
    end)

    -- 2) adiciona coluna faixa_etaria baseada em "idade"
    local idade_col = cleaned["idade"]
    if idade_col ~= nil then
        local faixa_col = idade_col:apply(function(v)
            return faixa_etaria_from_idade(v)
        end)
        cleaned:new_column("faixa_etaria", faixa_col)
    end

    -- exporta como payload "clientes"
    return {
        clientes = cleaned
    }
end
