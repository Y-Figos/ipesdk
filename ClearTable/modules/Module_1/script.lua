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

function main()
    -- dataframe de entrada injetado pelo engine
    local df = Module_1_input

    -- Filtra linhas inválidas
    local cleaned = df:filter(function(row)
        -- colunas: id, nome, idade, email, endereco
        local nome     = row["nome"]
        local idade    = row["idade"]
        local email    = row["email"]
        local endereco = row["endereco"]

        -- Nome obrigatório
        if not is_nonempty(nome) then
            return false
        end

        -- Email obrigatório e com '@'
        if not is_nonempty(email) or not contains_at(email) then
            return false
        end

        -- Idade numérica e plausível
        if not is_valid_age(idade) then
            return false
        end

        -- Endereço válido
        if not is_valid_address(endereco) then
            return false
        end

        return true
    end)

    return {
        output = cleaned
    }
end

